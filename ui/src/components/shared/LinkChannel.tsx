import {Channel, GuildId} from "../../data/discord";

export const LinkChannel = (props: {
    channel: Channel,
    children?: JSX.Element
}) => {
    const url = `https://discord.com/channels/${GuildId}/${props.channel.Id}`;
    const children = props.children ? props.children : props.channel.Name;
    return <a href={url}>{children}</a>;
};
