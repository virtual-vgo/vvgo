import {Credit} from "./Credit";
import {Director} from "./Director";
import {GuildMember} from "./GuildMember";
import {Highlight} from "./Highlight";
import {MixtapeProject} from "./MixtapeProject";
import {Part} from "./Part";
import {Project} from "./Project";
import {Session} from "./Session";

export const Endpoint = "/api/v1";

export enum ApiStatus {
    Ok = "ok",
    Error = "error"
}

export type ApiDataset = Highlight[] | Director[] | Credit[]

export interface ApiResponse {
    Status: string;
    Error: ErrorResponse;
    Dataset: ApiDataset;
    Parts: Part[];
    Projects: Project[];
    Sessions: Session[];
    Identity: Session;
    GuildMembers: GuildMember[];
    MixtapeProjects: MixtapeProject[];
    OAuthRedirect: OAuthRedirect;
}

export interface OAuthRedirect {
    DiscordURL: string;
    State: string;
    Secret: string;
}

export interface ErrorResponse {
    Code: number;
    Error: string;
}
