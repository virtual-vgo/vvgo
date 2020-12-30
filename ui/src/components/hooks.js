import {useEffect, useState} from 'react';

const axios = require('axios').default;

export function useLoginRoles() {
    const params = new URLSearchParams(window.location.search)
    const url = '/roles?' + params.getAll("roles").map(value => 'roles=' + value).join('&')
    return useAndCacheApiData(url, [])
}

export function useProjects() {
    return useAndCacheApiData('/projects_api', [])
}

export function useParts() {
    return useAndCacheApiData('/parts_api', [])
}

export const StatusNeedsRun = 'statusNeedsRun'
export const StatusRunning = 'statusRunning'
export const StatusComplete = 'statusComplete'
export const StatusFailure = 'statusFailure'

export function useAndCacheApiData(url, initialState) {
    const [data, setData] = useState(initialState)
    const [status, setStatus] = useState(StatusRunning)
    const [cachedUrl, setCachedUrl] = useState(null)
    useEffect(() => {
        if (status === StatusNeedsRun || cachedUrl !== url) {
            setStatus(StatusRunning)
            axios.get(url)
                .then(response => {
                    setData(response.data)
                    setCachedUrl(url)
                    setStatus(StatusComplete)
                })
                .catch(error => {
                    console.log(error)
                    setStatus(StatusFailure)
                })
        }
    }, [url, status, cachedUrl]);
    return {data, status, setStatus}
}
