import {useEffect,useState} from 'react';
export function navigate(path:string){history.pushState({},'',path);window.dispatchEvent(new PopStateEvent('popstate'))}
export function useRoute(){const [route,setRoute]=useState(location.pathname+location.search);useEffect(()=>{const update=()=>setRoute(location.pathname+location.search);addEventListener('popstate',update);return()=>removeEventListener('popstate',update)},[]);return route;}
