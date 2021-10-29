export interface GuildMember {
    user: DiscordUser;
    nick: string;
    roles: string[];
}

export interface DiscordUser {
    id: string;
}
