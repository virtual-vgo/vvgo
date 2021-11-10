import { Channel, GuildId } from "../../data/discord";
import { GuildMember } from "../../datasets";

export const LinkChannel = (props: {
  channel: Channel;
  children?: JSX.Element;
}) => {
  const url = `https://discord.com/channels/${GuildId}/${props.channel.Id}`;
  const children = props.children ? props.children : props.channel.Name;
  return <a href={url}>{children}</a>;
};

export const LinkUser = (props: {
  member: GuildMember;
  children?: JSX.Element;
}) => {
  const url = `https://discord.com/users/${props.member.user.id}`;
  return <a href={url}>{props.children ?? props.member.displayName()}</a>;
};
