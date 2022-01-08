import { get } from "lodash/fp";

export class DiscordUser {
  id = "";
  username = "";

  static fromApiObject(obj: object): DiscordUser {
    const user = new DiscordUser();
    user.id = get("id", obj);
    user.username = get("username", obj);
    return user;
  }
}
