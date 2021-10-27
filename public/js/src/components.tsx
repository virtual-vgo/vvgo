import {Project} from "./datasets";
import React = require("react");

export const RootContainer = (props: { children: JSX.Element | JSX.Element[] }) =>
    <div className={"mt-4 container"}>{props.children}</div>;

// export const Visualization = (props: { drawSVG: (arg0: d3.Selection<SVGSVGElement, unknown, HTMLElement, any>, arg1: any) => void; }) => {
//     const ref = useRef();
//     const div = <div ref={ref}/>;
//     const svg = d3.select(ref.current).append("svg");
//     useEffect((): any => {
//         props.drawSVG(svg, props);
//         return () => svg.remove();
//     });
//     return div;
// };

export const GeekSquad = (props: { children?: JSX.Element }) => {
    const geekSquadChannel = "https://discord.com/channels/690626216637497425/691857421437501472";
    const children = props.children ? props.children : "#geek-squad";
    return <a href={geekSquadChannel}>{children}</a>;
};

export const YoutubeIframe = (props: { latest: Project }) => {
    const latest = props.latest;
    if (latest === undefined) return <div/>;
    if (latest.YoutubeEmbed === "") return <div/>;
    return <div className="project-iframe-wrapper text-center m-2">
        <iframe className="project-iframe" src={latest.YoutubeEmbed}
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                allowFullScreen/>
    </div>;
};
