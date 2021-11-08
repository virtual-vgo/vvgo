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
    error?: ApiError;
    creditsPasta?: CreditsPasta;
    creditsTable?: CreditsTable;
    dataset?: Dataset;
    guildMembers?: GuildMember[];
    identity?: Session;
    mixtapeProjects?: MixtapeProject[];
    oauthRedirect?: OAuthRedirect;
    parts?: Part[];
    projects?: Project[];
    sessions?: Session[];

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
                apiResp.creditsPasta = CreditsPasta.fromApiObject(get("CreditsPasta", obj));
                apiResp.dataset = Dataset.fromApiObject(get("Dataset", obj));
                apiResp.identity = Session.fromApiObject(get("Identity", obj));
                apiResp.oauthRedirect = OAuthRedirect.fromApiObject(get("OAuthRedirect", obj));

                apiResp.creditsTable = get("CreditsTable", obj) as CreditsTable;
                apiResp.guildMembers = get("GuildMembers", obj)?.map((p: object[]) => GuildMember.fromApiObject(p));
                apiResp.mixtapeProjects = get("MixtapeProjects", obj)?.map((p: object[]) => MixtapeProject.fromApiObject(p));
                apiResp.parts = get("Parts", obj)?.map((p: object[]) => Part.fromApiObject(p));
                apiResp.projects = get("Projects", obj)?.map((p: object[]) => Project.fromApiObject(p));
                apiResp.sessions = get("Sessions", obj)?.map((p: object[]) => Session.fromApiObject(p));
                break;

            default:
                throw `invalid api response`;
        }
        return apiResp;
    }
}

