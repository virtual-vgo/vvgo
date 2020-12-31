import {useEffect, useState} from 'react';

export function useDrawerState(initialState) {
    const [state, setState] = useState(initialState);
    return {isOpen: state, openDrawer: () => setState(true), closeDrawer: () => setState(false)}
}

export function useLoginRoles() {
    return useAndCacheApiData('/api/v1/roles', [])
}

export function useProjects() {
    return useAndCacheApiData('/api/v1/projects', [])
}

export function useParts() {
    return useAndCacheApiData('/api/v1/parts', [])
}

export function useLeaders() {
    return useAndCacheApiData('/api/v1/leaders', [])
}

export const Status = Object.freeze({
    NeedsRun: 'needsRun',
    Running: 'running',
    Complete: 'complete',
    Failure: 'failure'
})

export function useAndCacheApiData(url, initialState) {
    const [data, setData] = useState(initialState)
    const [status, setStatus] = useState(Status.NeedsRun)
    const [cachedUrl, setCachedUrl] = useState(null)
    useEffect(() => {
        if (status === Status.NeedsRun || cachedUrl !== url) {
            setStatus(Status.Running)
            fetch(url)
                .then(response => response.json())
                .then(data => {
                    setData(data)
                    setCachedUrl(url)
                    setStatus(Status.Complete)
                })
                .catch(error => {
                    console.log(error)
                    setStatus(Status.Failure)
                })
        }
    }, [data, status, url, cachedUrl]);
    return {data, setData, status, setStatus}
}
