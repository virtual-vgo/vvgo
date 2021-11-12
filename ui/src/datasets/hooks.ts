import isEmpty from "lodash/fp/isEmpty";
import { useEffect, useState } from "react";
import { getSession } from "../auth";
import { ApiError } from "./ApiError";
import { ApiResponse, ApiStatus, Endpoint } from "./ApiResponse";
import { Credit } from "./Credit";
import { CreditsTable } from "./CreditsTable";
import { DatasetRow } from "./Dataset";
import { Director } from "./Director";
import { GuildMember } from "./GuildMember";
import { Highlight } from "./Highlight";
import { Part } from "./Part";
import { Project } from "./Project";
import { ApiRole, Session, SessionKind } from "./Session";

export const useCredits = (): Credit[] | undefined =>
  useDataset("Credits", Credit.fromDatasetRow);

export const useCreditsTable = (project: Project): CreditsTable | undefined => {
  const params = new URLSearchParams({ projectName: project.Name });
  const url = "/credits/table?" + params.toString();
  return useApiData(url, (p) => p.creditsTable ?? []);
};

export const useDirectors = (): Director[] | undefined =>
  useDataset("Leaders", Director.fromDatasetRow);

export const useGuildMembers = (): GuildMember[] | undefined =>
  useApiData("/guild_members/list", (resp) => resp.guildMembers);

export const useGuildMemberSearch = (
  query: string,
  limit: number
): GuildMember[] => {
  const [data, setData] = useState({} as ApiResponse);
  const params = new URLSearchParams({ query: query, limit: limit.toString() });
  const url = `/guild_members/search?` + params.toString();
  useEffect(() => {
    isEmpty(query)
      ? setData({} as ApiResponse)
      : fetchApi(url, { method: "GET" }).then((resp) => setData(resp));
  }, [url]);
  return data.guildMembers ?? [];
};

export const useGuildMemberLookup = (ids: string[]) => {
  const [data, setData] = useState({} as ApiResponse);
  const url = `/guild_members/lookup?`;
  useEffect(() => {
    if (ids.length > 0)
      fetchApi(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(ids),
      }).then((resp) => setData(resp));
  }, [url, ids.sort().join(",")]);
  return data.guildMembers ?? [];
};

export const useHighlights = (): Highlight[] | undefined =>
  useDataset("Highlights", Highlight.fromDatasetRow);

export const useParts = (): Part[] | undefined =>
  useApiData("/parts", (resp) => resp.parts);
export const useProjects = (): Project[] | undefined =>
  useApiData("/projects", (resp) => resp.projects);

export const useSessions = (): [
  Session[] | undefined,
  (sessions: Session[]) => void
] => useAndSetApiData("/sessions", (resp) => resp.sessions);

export const useNewApiSession = (
  expires: number,
  roles: Array<ApiRole>
): Session | undefined => {
  const [session, setSession] = useState<Session | undefined>(undefined);
  useEffect(() => {
    Session.Create(SessionKind.ApiToken, roles, expires).then((resp) =>
      setSession(resp.sessions?.pop())
    );
  }, [roles.toString()]);
  return session;
};

export function useDataset<T>(
  name: string,
  parseRow: (x: DatasetRow) => T
): T[] | undefined {
  return useApiData("/dataset?name=" + name, (p) =>
    p.dataset?.map((row) => parseRow(row))
  );
}

export function useApiData<T>(
  url: RequestInfo,
  getData: (r: ApiResponse) => T
): T | undefined {
  const [data] = useAndSetApiData(url, getData);
  return data;
}

export function useAndSetApiData<T>(
  url: RequestInfo,
  getData: (r: ApiResponse) => T
): [T | undefined, (t: T | undefined) => void] {
  const [data, setData] = useState<T | undefined>(undefined);
  useEffect(() => {
    fetchApi(url, { method: "GET" }).then((resp) => setData(getData(resp)));
  }, [url, getSession().key]);
  return [data, setData];
}

export const fetchApi = async (
  url: RequestInfo,
  init: RequestInit
): Promise<ApiResponse> => {
  init.headers = {
    ...init.headers,
    Authorization: "Bearer " + getSession().key,
  };
  console.log("Api Request:", init.method, url);
  return fetch(Endpoint + url, init)
    .then((resp) => resp.json())
    .then((respJson) => {
      const resp = ApiResponse.fromApiJSON(respJson);
      console.log("Api Response:", resp);
      if (resp.status == ApiStatus.Error) {
        throw resp.error ?? ApiError.UnknownError;
      }
      return resp;
    });
};
