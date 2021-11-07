import {Credit} from "./Credit";
import {CreditsTable} from "./CreditsTable";
import {Director} from "./Director";
import {GuildMember} from "./GuildMember";
import {Highlight} from "./Highlight";
import {mixtapeProject} from "./mixtapeProject";
import {Part} from "./Part";
import {Project} from "./Project";
import {Session} from "./Session";

export const Endpoint = "/api/v1";

export enum ApiStatus {
    Ok = "ok",
    Error = "error",
    Found = "found"
}

export type ApiDataset = Highlight[] | Director[] | Credit[]

export interface ApiResponse {
    Status: string;
    Error?: ErrorResponse;
    Dataset?: ApiDataset;
    Parts?: Part[];
    Projects?: Project[];
    Sessions?: Session[];
    Identity?: Session;
    GuildMembers?: GuildMember[];
    MixtapeProjects?: mixtapeProject[];
    OAuthRedirect?: OAuthRedirect;
    CreditsTable?: CreditsTable;
    CreditsPasta?: CreditsPasta;
    Spreadsheet?: Spreadsheet;
}

export interface Spreadsheet {
    SpreadsheetName: string;
    sheets?: Sheet[];
}

export interface Sheet {
    Name: string;
    Values: string[][];
}

export interface CreditsPasta {
    WebsitePasta: string;
    VideoPasta: string;
    YoutubePasta: string;
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
