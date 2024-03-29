import { get } from "lodash/fp";
import { ApiResponse } from "./ApiResponse";
import { GuildMember } from "./GuildMember";
import { fetchApi } from "./hooks";

export class MixtapeProject {
  id = 0;
  Name = "";
  mixtape = "";
  title = "";
  blurb = "";
  channel = "";
  hosts: string[] = [];

  static fromApiObject(obj: object): MixtapeProject {
    const project = new MixtapeProject();
    project.id = get("id", obj);
    project.Name = get("Name", obj);
    project.mixtape = get("mixtape", obj);
    project.Name = get("Name", obj);
    project.title = get("title", obj) ?? "";
    project.blurb = get("blurb", obj) ?? "";
    project.channel = get("channel", obj) ?? "";
    project.hosts = get("hosts", obj) ?? [];
    return project;
  }

  resolveNicks(members: GuildMember[]): string[] {
    return members
      .filter((m) => this.hosts?.includes(m.user.id))
      .map((m) => m.displayName());
  }

  toApiObject(): object {
    return {
      Name: this.Name,
      mixtape: this.mixtape,
      title: this.title,
      blurb: this.blurb,
      channel: this.channel,
      hosts: this.hosts,
    };
  }

  create(): Promise<ApiResponse> {
    return fetchApi(`/mixtape/projects/`, {
      method: "POST",
      body: JSON.stringify(this.toApiObject()),
    });
  }

  save(): Promise<ApiResponse> {
    return fetchApi(`/mixtape/projects/${this.id}`, {
      method: "PUT",
      body: JSON.stringify(this.toApiObject()),
    });
  }

  delete(): Promise<ApiResponse> {
    return fetchApi(`/mixtape/projects/${this.id}`, { method: "DELETE" });
  }
}
