import { get } from "lodash/fp";
import { ApiError } from "./ApiError";
import { CreditsPasta } from "./CreditsPasta";
import { CreditsTable } from "./CreditsTable";
import { Dataset } from "./Dataset";
import { GuildMember } from "./GuildMember";
import { MixtapeProject } from "./MixtapeProject";
import { OAuthRedirect } from "./OAuthRedirect";
import { Part } from "./Part";
import { Project } from "./Project";
import { Session } from "./Session";

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
  mixtapeProject?: MixtapeProject;
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
        apiResp.error = ApiError.fromApiObject(get("Error", obj));
        break;

      case ApiStatus.Ok:
        apiResp.creditsPasta = CreditsPasta.fromApiObject(
          get("CreditsPasta", obj)
        );
        apiResp.creditsTable = CreditsTable.fromApiArray(
          get("CreditsTable", obj)
        );
        apiResp.dataset = Dataset.fromApiArray(get("Dataset", obj));
        apiResp.guildMembers = get("GuildMembers", obj)?.map((p: object[]) =>
          GuildMember.fromApiObject(p)
        );
        apiResp.identity = Session.fromApiObject(get("Identity", obj));
        apiResp.mixtapeProject = MixtapeProject.fromApiObject(
          get("MixtapeProject", obj)
        );
        apiResp.mixtapeProjects = get("MixtapeProjects", obj)?.map(
          (p: object[]) => MixtapeProject.fromApiObject(p)
        );
        apiResp.oauthRedirect = OAuthRedirect.fromApiObject(
          get("OAuthRedirect", obj)
        );
        apiResp.parts = get("Parts", obj)?.map((p: object[]) =>
          Part.fromApiObject(p)
        );
        apiResp.projects = get("Projects", obj)?.map((p: object[]) =>
          Project.fromApiObject(p)
        );
        apiResp.sessions = get("Sessions", obj)?.map((p: object[]) =>
          Session.fromApiObject(p)
        );
        break;

      default:
        throw `invalid api response`;
    }
    return apiResp;
  }
}
