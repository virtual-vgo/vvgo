export class GuildMember {
  user: DiscordUser = new DiscordUser();
  nick = "";
  roles: string[] = [];

  static fromApiObject(obj: object): GuildMember {
    return obj as GuildMember;
  }
}

export class DiscordUser {
  id = "";
}
