type BrandMarkProps = { size?: number; className?: string };

export default function BrandMark({size = 42, className = ''}: BrandMarkProps) {
  return (
    <svg
      className={`brand-mark ${className}`.trim()}
      width={size}
      height={size}
      viewBox="0 0 48 48"
      role="img"
      aria-label="HIVEPlace"
    >
      <path d="M24 3 42.2 13.5v21L24 45 5.8 34.5v-21L24 3Z" fill="none" stroke="currentColor" strokeWidth="3.2" />
      <path d="M24 10.2 36 17.1v13.8L24 37.8 12 30.9V17.1L24 10.2Z" fill="none" stroke="currentColor" strokeWidth="3.2" />
      <path d="m24 17.4 5.8 3.3v6.6L24 30.6l-5.8-3.3v-6.6l5.8-3.3Z" fill="currentColor" />
    </svg>
  );
}
