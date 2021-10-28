import React = require("react");

export const GeekSquad = (props: { children?: JSX.Element }) => {
    const geekSquadChannel = "https://discord.com/channels/690626216637497425/691857421437501472";
    const children = props.children ? props.children : "#geek-squad";
    return <a href={geekSquadChannel}>{children}</a>;
};
