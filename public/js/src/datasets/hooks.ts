import {ApiResponse, ApiResponseStatus, Endpoint, ErrorResponse} from "./api_response";
import {Session} from "./session";
import {useEffect, useState} from "react";
import {Credit} from "./credit";
import {Director} from "./director";
import {Highlight} from "./highlight";
import {Part} from "./part";
import {Project} from "./project";
import _ = require('lodash');

export const useCredits = (): Credit[] => useDataset("Credits")
export const useDirectors = (): Director[] => useDataset("Leaders")
export const useHighlights = (): Highlight[] => useDataset("Highlights")
export const useMySession = (): Session => useApiData("/me", "Identity", new Session())
export const useParts = (): Part[] => useApiData("/parts", "Parts", [])
export const useProjects = (): Project[] => useApiData("/projects", "Projects", [])
export const useSessions = (): [Session[], (sessions: Session[]) => void] =>
    useAndSetApiData("/sessions", "Sessions", [])

export function useDataset<T>(name: string): T[] {
    return useApiData("/dataset?name=" + name, "Dataset", [])
}

function useApiData<T>(url: RequestInfo, key: string, defaultValue: T): T {
    const [data] = useAndSetApiData(url, key, defaultValue)
    return data as T
}

function useAndSetApiData<T>(url: RequestInfo, key: string, defaultValue: T): [T, (t: T) => void] {
    const [data, setData] = useState(new ApiResponse())
    useEffect(() => {
        fetchApi(url, {method: 'GET'}).then(resp => setData(resp))
    }, [url])
    return [_.get(data, key, defaultValue) as T, (t: T) => setData(_.set(data, key, t))]
}

export const fetchApi = async (url: RequestInfo, init: RequestInit): Promise<ApiResponse> => {
    console.log("Api Request:", init.method, url)
    return fetch(Endpoint + url, init)
        .then(response => response.json())
        .then(obj => {
            const resp = obj as ApiResponse
            console.log("Api Response:", resp)
            if (resp.Status === ApiResponseStatus.Error) {
                const error = _.get(ApiResponse, "Error", {Error: "unknown", Code: 0}) as ErrorResponse
                throw `vvgo error [${error.Code}]: ${error.Error}`
            }
            return resp
        })
}
