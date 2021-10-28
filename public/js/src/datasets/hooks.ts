import {useEffect, useState} from "react";
import {getSession} from "../auth";
import {ApiDataset, ApiResponse, ApiStatus, Endpoint, ErrorResponse} from "./ApiResponse";
import {Credit} from "./Credit";
import {Director} from "./Director";
import {GuildMember} from "./GuildMember";
import {Highlight} from "./Highlight";
import {MixtapeProject} from "./MixtapeProject";
import {Part} from "./Part";
import {Project} from "./Project";
import {Session} from "./Session";
import _ = require("lodash");

export const useCredits = (): Credit[] => useDataset("Credits");
export const useDirectors = (): Director[] => useDataset("Leaders");
export const useGuildMembers = (query: string, limit: number): GuildMember[] => {
    const [data, setData] = useState({} as ApiResponse);
    const url = `/guild_members?query=${query}&limit=${limit}`;
    useEffect(() => {
        if (query !== "")
            fetchApi(url, {method: "GET"}).then(resp => setData(resp));
    }, [url]);

    return data.GuildMembers ? data.GuildMembers : [] as GuildMember[];
};
export const useHighlights = (): Highlight[] => useDataset("Highlights");
export const useMixtapeProjects = (): [MixtapeProject[], (projects: MixtapeProject[]) => void] =>
    useAndSetApiData("/mixtape", (p) => _.defaultTo(p.MixtapeProjects, []));
export const useParts = (): Part[] =>
    useApiData("/parts", (p) => _.defaultTo(p.Parts, []));
export const useProjects = (): Project[] =>
    useApiData("/projects", (p) => _.defaultTo(p.Projects, []));
export const useSessions = (): [Session[], (sessions: Session[]) => void] =>
    useAndSetApiData("/sessions", (p) => _.defaultTo(p.Sessions, []));

export function useDataset<T extends ApiDataset>(name: string): T {
    return useApiData("/dataset?name=" + name, (p) => _.defaultTo(p.Dataset, [])) as T;
}

export function useApiData<T>(url: RequestInfo, getData: (r: ApiResponse) => T): T {
    const [data] = useAndSetApiData(url, getData);
    return data as T;
}

export function useAndSetApiData<T>(url: RequestInfo, getData: (r: ApiResponse) => T): [T, (t: T) => void] {
    const [data, setData] = useState(null as T);
    useEffect(() => {
        fetchApi(url, {method: "GET"}).then(resp => setData(getData(resp)));
    }, [url, getSession().Key]);
    return [data, setData];
}

export const fetchApi = async (url: RequestInfo, init: RequestInit): Promise<ApiResponse> => {
    init.headers = {...init.headers, Authorization: "Bearer " + getSession().Key};
    console.log("Api Request:", init.method, url);
    return fetch(Endpoint + url, init)
        .then(response => response.json())
        .then(obj => {
            const resp = obj as ApiResponse;
            console.log("Api Response:", resp);
            if (resp.Status === ApiStatus.Error) {
                const error = _.get(resp, "Error", {Error: "unknown", Code: 0}) as ErrorResponse;
                throw `vvgo error [${error.Code}]: ${error.Error}`;
            }
            return resp;
        });
};
