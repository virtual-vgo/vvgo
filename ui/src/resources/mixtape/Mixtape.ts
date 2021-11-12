import { Client } from "../../clients/vvgo";
import { Resource } from "../base/Resource";
import { Projects } from "./Projects";

export class Mixtape extends Resource {
  projects: Projects;

  constructor(client: Client) {
    super(client);
    this.projects = new Projects(client);
  }
}
