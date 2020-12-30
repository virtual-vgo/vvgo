import {useEffect, useState} from 'react';

const axios = require('axios').default;

export function useLoginRoles() {
    return useAndCacheApiData('/roles', [])
}

export function useProjects() {
    return useAndCacheApiData('/projects_api', [])
}

export function useParts() {
    return useAndCacheApiData('/parts_api', [])
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
            axios.get(url)
                .then(response => {
                    setData(response.data)
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
