import {useEffect, useState} from "react";
import {getSession} from "../auth";
import {ApiDataset, ApiResponse, ApiStatus, Endpoint, ErrorResponse, Sheet, Spreadsheet} from "./ApiResponse";
import {Credit} from "./Credit";
import {CreditsTable} from "./CreditsTable";
import {Director} from "./Director";
import {GuildMember} from "./GuildMember";
import {Highlight} from "./Highlight";
import {mixtapeProject} from "./MixtapeProject";
import {Part} from "./Part";
import {Project} from "./Project";
import {ApiRole, createSessions, Session, SessionKind} from "./Session";
import _ = require("lodash");

export const useCredits = (): Credit[] => useDataset("Credits");

export const useCreditsTable = (project: Project): CreditsTable => {
    const params = new URLSearchParams({projectName: project.Name});
    const url = "/credits/table?" + params.toString();
    return useApiData(url, (p) => _.defaultTo(p.CreditsTable, []));
};

export const useDirectors = (): Director[] => useDataset("Leaders");

export const useGuildMemberSearch = (query: string, limit: number): GuildMember[] => {
    const [data, setData] = useState({} as ApiResponse);
    const params = new URLSearchParams({query: query, limit: limit.toString()});
    const url = `/guild_members/search?` + params.toString();
    useEffect(() => {
        _.isEmpty(query) ?
            setData({} as ApiResponse) :
            fetchApi(url, {method: "GET"}).then(resp => setData(resp));
    }, [url]);
    return _.defaultTo(data.GuildMembers, []);
};

export const useGuildMemberLookup = (ids: string[]) => {
    const [data, setData] = useState({} as ApiResponse);
    const url = `/guild_members/lookup?`;
    useEffect(() => {
        if (ids.length > 0) fetchApi(url, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(ids),
        }).then(resp => setData(resp));
    }, [url, ids.sort().join(",")]);
    return _.defaultTo(data.GuildMembers, []);
};

export const useHighlights = (): Highlight[] => useDataset("Highlights");

export const useMixtapeProjects = (): [mixtapeProject[], (projects: mixtapeProject[]) => void] => {
    const [projects, setProjects] = useAndSetApiData("/mixtape/projects", (r) =>
        _.defaultTo(r.MixtapeProjects, []));
    return [_.defaultTo(projects, []).map(p => ({
        ...p,
        tags: _.defaultTo(p.tags, []),
        hosts: _.defaultTo(p.hosts, []),
    })), setProjects];
};

export const useParts = (): Part[] =>
    useApiData("/parts", (p) => _.defaultTo(p.Parts, []));

export const useProjects = (): Project[] =>
    useApiData("/projects", (p) => _.defaultTo(p.Projects, []));

export const useSessions = (): [Session[], (sessions: Session[]) => void] =>
    useAndSetApiData("/sessions", (p) => _.defaultTo(p.Sessions, []));

export const useNewApiSession = (expires: number, roles: Array<ApiRole>): Session => {
    const [session, setSession] = useState(null as Session);
    useEffect(() => {
        createSessions([{expires: expires, Kind: SessionKind.ApiToken, Roles: roles}])
            .then(sessions => _.isEmpty(sessions) ? setSession({} as Session) : setSession(sessions[0]));
    }, [roles.toString()]);
    return session;
};

export const useSheet = (spreadsheetName: string, sheetName: string): Sheet =>
    _.defaultTo(_.defaultTo(useSpreadsheet(spreadsheetName, [sheetName]), {} as Spreadsheet).sheets, [])
        .filter(sheet => sheet.Name == sheetName).pop();

export const useSpreadsheet = (spreadsheetName: string, sheetNames: string[]): Spreadsheet => {
    const params = new URLSearchParams({spreadsheetName: spreadsheetName, sheetNames: sheetNames.join(",")});
    return useApiData("/spreadsheet?" + params.toString(), (p) => _.defaultTo(p.Spreadsheet, {} as Spreadsheet));
};

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
