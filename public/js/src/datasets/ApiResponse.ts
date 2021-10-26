import {Credit} from "./Credit";
import {Director} from "./Director";
import {GuildMember} from "./guildMember";
import {Highlight} from "./Highlight";
import {MixtapeProject} from "./MixtapeProject";
import {Part} from "./part";
import {Project} from "./project";
import {Session} from "./session";

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
}

export interface ErrorResponse {
    Code: number;
    Error: string;
}
