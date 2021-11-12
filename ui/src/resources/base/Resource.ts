import { Client } from "../../clients/vvgo";

export class Resource {
  readonly client: Client;

  constructor(client: Client) {
    this.client = client;
  }
}
