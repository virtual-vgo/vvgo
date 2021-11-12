import { GuildMember } from "../../datasets";
import { ProjectSchema as ProjectSchema } from "../schema/mixtape/ProjectSchema";
import { CreateProjectArgs, SaveProjectArgs } from "./Projects";

export class Project {
  id = 0;
  Name = "";
  mixtape = "";
  title = "";
  blurb = "";
  channel = "";
  hosts: string[] = [];

  static fromApiObject(obj: ProjectSchema | undefined): Project {
    const project = new Project();
    project.id = obj?.id ?? 0;
    project.mixtape = obj?.mixtape ?? "";
    project.Name = obj?.Name ?? "";
    project.title = obj?.title ?? "";
    project.blurb = obj?.blurb ?? "";
    project.channel = obj?.channel ?? "";
    project.hosts = obj?.hosts ?? [];
    return project;
  }

  createArgs(): CreateProjectArgs {
    return {
      Name: this.Name,
      mixtape: this.mixtape,
      title: this.title,
      blurb: this.blurb,
      channel: this.channel,
      hosts: this.hosts,
    };
  }

  saveArgs(): SaveProjectArgs {
    return this.createArgs();
  }

  resolveNicks(members: GuildMember[]): string[] {
    return members
      .filter((m) => this.hosts?.includes(m.user.id))
      .map((m) => m.displayName());
  }
}
