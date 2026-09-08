#!/usr/bin/env python3
"""Reproduce the pre-COMMIT deadline hole in isolated WAL-logged SQL objects.

Run from repository root. No production tables or global settings are changed.
The deferred trigger delays COMMIT processing; it does not simulate disk fsync.
"""
import datetime
import json
import pathlib
import subprocess
import uuid

root = pathlib.Path(__file__).resolve().parents[4]
schema = "r2_deadline_probe_" + uuid.uuid4().hex
sql = f"""
CREATE SCHEMA {schema};
SET search_path TO {schema};
SET synchronous_commit TO on;
CREATE TABLE candidates(id int PRIMARY KEY, deadline timestamptz NOT NULL,
  checked_at timestamptz NOT NULL, representation text NOT NULL);
CREATE TABLE probes(label text, observation jsonb);
CREATE FUNCTION delay_commit() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN PERFORM pg_sleep(0.4); RETURN NEW; END $$;
CREATE CONSTRAINT TRIGGER slow_commit AFTER INSERT ON candidates
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION delay_commit();
BEGIN;
INSERT INTO candidates SELECT 1,clock_timestamp()+interval '150 milliseconds',
  clock_timestamp(),'{{"status":"SUCCEEDED","marker":"real-candidate"}}';
COMMIT;
INSERT INTO probes SELECT 'precommit_check_counterexample',jsonb_build_object(
  'check_was_before_deadline',checked_at < deadline,
  'postcommit_observation_after_deadline',clock_timestamp() >= deadline,
  'candidate_visible',true,'checked_at',checked_at,'deadline',deadline,
  'postcommit_observed_at',clock_timestamp(),
  'unsafe_algorithm_would_publish_success',checked_at < deadline)
FROM candidates WHERE id=1;
ALTER TABLE candidates DISABLE TRIGGER slow_commit;
BEGIN;
INSERT INTO candidates SELECT 2,clock_timestamp()+interval '150 milliseconds',
  clock_timestamp(),'{{"status":"SUCCEEDED","marker":"early-candidate"}}';
COMMIT;
INSERT INTO probes SELECT 'early_candidate_ack',jsonb_build_object(
  'ack_upper_bound_before_deadline',clock_timestamp()<deadline,
  'candidate_id',id,'deadline',deadline,'ack_upper_bound',clock_timestamp())
FROM candidates WHERE id=2;
SELECT pg_sleep(0.25);
INSERT INTO probes SELECT 'late_promotion_counterexample',jsonb_build_object(
  'candidate_ack_before_deadline',
    (SELECT (observation->>'ack_upper_bound_before_deadline')::boolean FROM probes WHERE label='early_candidate_ack'),
  'promotion_after_deadline',clock_timestamp() >= deadline,
  'would_be_first_public_availability_after_deadline',true)
FROM candidates WHERE id=2;
INSERT INTO probes VALUES ('uncertainty_interval',jsonb_build_object(
  'sample_ms',95,'uncertainty_ms',10,'deadline_ms',100,
  'point_comparison_accepts',95<100,'conservative_upper_bound_accepts',95+10<100,
  'equality_accepts',100<100,
  'unknown_uncertainty_accepts',false));
SELECT jsonb_build_object('postgres',version(),'fsync',current_setting('fsync'),
 'synchronous_commit',current_setting('synchronous_commit'),
 'wal_level',current_setting('wal_level'),'cases',jsonb_object_agg(label,observation)) FROM probes;
DROP SCHEMA {schema} CASCADE;
"""
command = ["docker", "compose", "-f", str(root / "hub/deploy/r2/compose.yaml"),
           "exec", "-T", "postgres", "psql", "-X", "-qAt", "-v", "ON_ERROR_STOP=1",
           "-U", "hub", "-d", "hub_core"]
result = subprocess.run(command, input=sql, text=True, capture_output=True, check=True)
record = json.loads(next(line for line in result.stdout.splitlines() if line.startswith('{')))
cases = record['cases']
assert cases['precommit_check_counterexample']['check_was_before_deadline']
assert cases['precommit_check_counterexample']['postcommit_observation_after_deadline']
assert cases['early_candidate_ack']['ack_upper_bound_before_deadline']
assert cases['late_promotion_counterexample']['promotion_after_deadline']
record.update({
    'date_utc': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'head': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip(),
    'worktree_has_uncommitted_changes': bool(subprocess.check_output(['git', 'status', '--porcelain'], cwd=root)),
    'requirements': ['EXE-11', 'R2-EXE-06', 'T-R2-01', 'IT-08'],
    'status': 'COUNTEREXAMPLES_REPRODUCED_NOT_REQUIREMENT_PASS',
    'limitations': ['Deferred trigger delays COMMIT processing, not physical disk fsync.',
                    'Clock uncertainty is an explicit synthetic interval, not host clock qualification.',
                    'No runtime domain code is exercised by this isolated SQL experiment.'],
})
print(json.dumps(record, indent=2))
