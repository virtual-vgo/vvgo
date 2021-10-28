import React = require("react");
import {Project} from "../../datasets";

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
