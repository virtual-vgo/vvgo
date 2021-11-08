import { get, isEmpty, uniq } from "lodash/fp";
import { ApiResponse } from "./ApiResponse";
import { GuildMember } from "./GuildMember";
import { fetchApi } from "./hooks";

export class MixtapeProject {
  Name: string;
  mixtape: string;
  title = "";
  blurb = "";
  channel = "";
  hosts: string[] = [];

  constructor(name: string, mixtape: string) {
    this.Name = name;
    this.mixtape = mixtape;
  }

  static fromApiObject(obj: object): MixtapeProject {
    const name = get("Name", obj);
    if (isEmpty(name)) throw `empty field name`;
    const mixtape = get("mixtape", obj);
    if (isEmpty(name)) throw `empty field mixtape`;
    const project = new MixtapeProject(name, mixtape);
    project.title = get("title", obj) ?? "";
    project.blurb = get("blurb", obj) ?? "";
    project.channel = get("channel", obj) ?? "";
    project.hosts = get("hosts", obj) ?? [];
    return project;
  }

  toApiObject(): object {
    return {
      Name: this.Name,
      Mixtape: this.mixtape,
      Title: this.title,
      Blurb: this.blurb,
      Channel: this.channel,
      Hosts: this.hosts,
    };
  }

  resolveNicks(members: GuildMember[]): string[] {
    return uniq(
      members.filter((m) => this.hosts?.includes(m.user.id)).map((m) => m.nick)
    );
  }

  save(): Promise<ApiResponse> {
    return fetchApi("/mixtape/projects", {
      method: "POST",
      body: JSON.stringify([this.toApiObject()]),
    });
  }

  delete(): Promise<ApiResponse> {
    return fetchApi("/mixtape/projects", {
      method: "DELETE",
      body: JSON.stringify([this.Name]),
    });
  }
}