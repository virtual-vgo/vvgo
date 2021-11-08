import {get, keys} from "lodash/fp";
import {ApiError} from "./ApiError";
import {CreditsPasta} from "./CreditsPasta";
import {CreditsTable} from "./CreditsTable";
import {Dataset} from "./Dataset";
import {GuildMember} from "./GuildMember";
import {MixtapeProject} from "./MixtapeProject";
import {OAuthRedirect} from "./OAuthRedirect";
import {Part} from "./Part";
import {Project} from "./Project";
import {Session} from "./Session";

export const Endpoint = "/api/v1";

export enum ApiStatus {
    Ok = "ok",
    Error = "error",
}

export class ApiResponse {
    status: ApiStatus;
    error: ApiError = new ApiError();
    creditsPasta: CreditsPasta = new CreditsPasta();
    creditsTable: CreditsTable = [];
    dataset: Dataset = new Dataset();
    guildMembers: GuildMember[] = [];
    identity: Session = new Session();
    mixtapeProjects: MixtapeProject[] = [];
    oauthRedirect: OAuthRedirect = new OAuthRedirect();
    parts: Part[] = [];
    projects: Project[] = [];
    sessions: Session[] = [];

    constructor(status: ApiStatus) {
        this.status = status;
    }

    static fromApiJSON(obj: object): ApiResponse {
        const apiResp: ApiResponse = new ApiResponse(get("Status", obj));
        switch (apiResp.status) {
            case ApiStatus.Error:
                apiResp.error = ApiError.fromApiJson(get("Error", obj));
                break;

            case ApiStatus.Ok:
                console.log(keys(obj));
                apiResp.creditsPasta = CreditsPasta.fromApiJSON(get("CreditsPasta", obj));
                apiResp.creditsTable = get("CreditsTable", obj) as CreditsTable;
                apiResp.dataset = Dataset.fromApiJSON(get("Dataset", obj));
                apiResp.guildMembers = get("GuildMembers", obj)?.map((p: object[]) => GuildMember.fromApiJSON(get("Dataset", p)));
                apiResp.identity = Session.fromApiObject(get("Identity", obj));
                apiResp.mixtapeProjects = get("MixtapeProjects", obj)?.map((p: object[]) => MixtapeProject.fromApiJSON(get("Dataset", p)));
                apiResp.oauthRedirect = OAuthRedirect.fromApiJSON(get("OAuthRedirect", obj));
                apiResp.parts = get("Parts", obj)?.map((p: object[]) => Part.fromApiJSON(get("Dataset", p)));
                apiResp.projects = get("Projects", obj)?.map((p: object[]) => Project.fromApiJSON(get("Dataset", p)));
                apiResp.sessions = get("Sessions", obj)?.map((p: object[]) => Session.fromApiObject(get("Dataset", p)));
                break;

            default:
                throw `invalid api response`;
        }
        return apiResp;
    }
}

