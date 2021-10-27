import {fetchApi} from "./hooks";

export enum Roles {
    ExecutiveDirector = "vvgo-leader",
    Member = "vvgo-member",
    Anonymous = "anonymous"
}

export enum SessionKind {
    Anonymous = "anonymous",
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

export const sessionIsAnonymous = (session: Session): boolean => {
    switch (true) {
        case session.Kind == SessionKind.Anonymous:
            return true;
        case session.Roles == undefined:
            return true;
        case session.Roles.length == 0:
            return true;
        case session.Roles.length == 1 && session.Roles[0] == Roles.Anonymous:
            return true;
        default:
            return false;
    }
};

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
