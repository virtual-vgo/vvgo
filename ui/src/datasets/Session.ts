import {isEmpty} from "lodash";
import {fetchApi} from "./hooks";

export enum UserRole {
    ExecutiveDirector = "vvgo-leader",
    VerifiedMember = "vvgo-member",
    ProductionTeam = "vvgo-teams",
    Anonymous = "anonymous"
}

export enum ApiRole {
    ReadSpreadsheet = "read_spreadsheet",
    WriteSpreadsheet = "write_spreadsheet",
    Download = "download",
}

export enum SessionKind {
    Anonymous = "anonymous",
    Bearer = "bearer",
    Basic = "basic",
    Discord = "discord",
    ApiToken = "api_token",
}

export interface Session {
    Kind: string;
    Key?: string;
    Roles?: string[];
    DiscordID?: string;
    CreatedAt?: string;
    ExpiresAt?: string;
}

export const AnonymousSession: Session = {Kind: SessionKind.Anonymous};

export const sessionIsAnonymous = (session: Session): boolean => {
    switch (true) {
        case session.Kind == SessionKind.Anonymous:
            return true;
        case isEmpty(session.Roles):
            return true;
        case session.Roles?.length == 1 && session.Roles[0] == UserRole.Anonymous :
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
