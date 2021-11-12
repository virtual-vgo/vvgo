import { Channel, GuildId } from "../../static/discord";
import { GuildMember } from "../../datasets";

export const LinkChannel = (props: {
  channel: Channel;
  children?: JSX.Element;
}) => {
  const url = `https://discord.com/channels/${GuildId}/${props.channel.Id}`;
  return <a href={url}>{props.children ?? props.channel.Name}</a>;
};

export const LinkUser = (props: {
  member: GuildMember;
  children?: JSX.Element;
}) => {
  const url = `https://discord.com/users/${props.member.user.id}`;
  return <a href={url}>{props.children ?? props.member.displayName()}</a>;
};
