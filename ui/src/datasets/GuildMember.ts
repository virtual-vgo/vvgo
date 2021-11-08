export class GuildMember {
    user: DiscordUser = new DiscordUser();
    nick: string = "";
    roles: string[] = [];

    static fromApiJSON(obj: object): GuildMember {
        return obj as GuildMember;
    }
}

export class DiscordUser {
    id: string = "";
}
