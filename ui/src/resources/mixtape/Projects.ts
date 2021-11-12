import { ApiResponse } from "../../datasets";
import { Resource } from "../base/Resource";
import { Project } from "./Project";

export interface CreateProjectArgs {
  Name?: string;
  mixtape?: string;
  title?: string;
  blurb?: string;
  channel?: string;
  hosts?: string[];
}

export type SaveProjectArgs = CreateProjectArgs;

export class Projects extends Resource {
  list(): Promise<Project[]> {
    return this.client
      .fetch("/mixtape/projects", { method: "GET" })
      .then((resp) => resp.mixtapeProjects ?? []);
  }

  create(project?: Project): Promise<Project> {
    if (project?.id != 0)
      throw `refusing to create a project with an existing id; did you mean save?`;
    return this.update("POST", undefined, project?.createArgs() ?? {});
  }

  save(project: Project): Promise<Project> {
    const id = project.id;
    if (id == 0) throw `id cannot be 0; did you mean create?`;
    return this.update("PUT", id, project.saveArgs());
  }

  delete(id: number): Promise<undefined> {
    if (id == 0) throw `id cannot be 0`;
    return this.client
      .fetch(`/mixtape/projects/${id}`, { method: "DELETE" })
      .then(() => undefined);
  }

  private update(
    method: "POST" | "PUT",
    id: number | undefined,
    args: CreateProjectArgs | SaveProjectArgs
  ): Promise<Project> {
    return this.client
      .fetch(`/mixtape/projects/${id ?? ""}`, {
        method: method,
        body: JSON.stringify(args),
      })
      .then((resp) => {
        if (!resp.mixtapeProject) throw `invalid api response`;
        return resp.mixtapeProject;
      });
  }
}
