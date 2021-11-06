export const GuildId = "690626216637497425";
export const GuildLink = "https://discord.gg/vvgo";

export interface Channel {
    Id: string;
    Name: string;
}

export const Channels = Object.freeze({
    GeekSquad: {Id: "691857421437501472", Name: "#geek-squad"} as Channel,
    NextProjectHints: {Id: "757726521837355159", Name: "#next-project-hints"} as Channel,
});

