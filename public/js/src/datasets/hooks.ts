import {ApiResponse, ApiStatuses, Endpoint, ErrorResponse} from "./ApiResponse";
import {Session} from "./session";
import {useEffect, useState} from "react";
import {Credit} from "./credit";
import {Director} from "./director";
import {Highlight} from "./highlight";
import {Part} from "./part";
import {Project} from "./project";
import {GuildMember} from "./guildMember";
import _ = require('lodash');

export const useCredits = (): Credit[] => useDataset("Credits")
export const useDirectors = (): Director[] => useDataset("Leaders")
export const useGuildMembers = (query: string, limit: number): GuildMember[] => {
    const [data, setData] = useState(new ApiResponse())

    const url = `/guild_members?query=${query}&limit=${limit}`
    useEffect(() => {
        if (query !== "")
            fetchApi(url, {method: 'GET'}).then(resp => setData(resp))
    }, [url])

    return data.GuildMembers ? data.GuildMembers : [] as GuildMember[]
}
export const useHighlights = (): Highlight[] => useDataset("Highlights")
export const useMixtapeProjects = (): MixtapeProject[] => useApiData("/mixtape", "MixtapeProjects", [])
export const useMySession = (): Session => useApiData("/me", "Identity", new Session())
export const useParts = (): Part[] => useApiData("/parts", "Parts", [])
export const useProjects = (): Project[] => useApiData("/projects", "Projects", [])
export const useSessions = (): [Session[], (sessions: Session[]) => void] =>
    useAndSetApiData("/sessions", "Sessions", [])

export function useDataset<T>(name: string): T[] {
    return useApiData("/dataset?name=" + name, "Dataset", [])
}

export function useApiData<T>(url: RequestInfo, key: string, defaultValue: T): T {
    const [data] = useAndSetApiData(url, key, defaultValue)
    return data as T
}

export function useAndSetApiData<T>(url: RequestInfo, key: string, defaultValue: T): [T, (t: T) => void] {
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
            if (resp.Status === ApiStatuses.Error) {
                const error = _.get(resp, "Error", {Error: "unknown", Code: 0}) as ErrorResponse
                throw `vvgo error [${error.Code}]: ${error.Error}`
            }
            return resp
        }).catch(err => {
            console.log(err)
            return new ApiResponse()
        })
}
