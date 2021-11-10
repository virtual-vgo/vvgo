import { get } from "lodash/fp";
import { DiscordUser } from "./DiscordUser";

export class GuildMember {
  user: DiscordUser = new DiscordUser();
  nick = "";
  roles: string[] = [];

  static fromApiObject(obj: object): GuildMember {
    const member = new GuildMember();
    member.user = DiscordUser.fromApiObject(get("user", obj));
    member.nick = get("nick", obj);
    member.roles = get("roles", obj);
    return member;
  }

  displayName(): string {
    if ((this.nick ?? "") != "") return this.nick;
    return this.user.username;
  }
}
