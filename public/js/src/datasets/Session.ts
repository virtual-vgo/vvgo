import {fetchApi} from "./hooks";

export enum SessionKind {
    Bearer = "bearer",
    Basic = "basic",
    Discord = "discord",
    ApiToken = "api_token",
}

export interface Session {
    Key: string;
    Kind: string;
    Roles: string[];
    DiscordID: string;
    ExpiresAt: string;
}

export const createSessions = async (sessions: { expires: number; Kind: SessionKind; Roles: string[] }[]) => {
    return fetchApi("/sessions", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(({"sessions": sessions})),
    }).then(resp => resp.Sessions);
};

export const deleteSessions = async (sessionsIds: string[]) => {
    return fetchApi("/sessions", {
        method: "DELETE",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(({"sessions": sessionsIds})),
    });
};
