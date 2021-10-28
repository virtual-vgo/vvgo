import React = require("react");
import {Project} from "../../datasets";

export const ProjectBanner = (props: { project: Project }) => {
    if (!props.project) return <div/>;
    const youtubeLink = props.project.YoutubeLink;
    const bannerLink = props.project.BannerLink;

    const Banner = () => {
        if (bannerLink === "") return <div>
            <h1 className="title">{props.project.Title}</h1>
            <h2>{props.project.Sources}</h2>
        </div>;
        else return <img src={bannerLink} className="mx-auto img-fluid" alt="banner"/>;
    };

    return <div id="banner" className={"d-flex justify-content-center"}>
        <a href={youtubeLink} className="text-light text-center">
            <Banner/>
        </a>
    </div>;
};
