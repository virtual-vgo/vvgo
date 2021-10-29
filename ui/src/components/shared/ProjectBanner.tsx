import React = require("react");
import {Project} from "../../datasets";

export const ProjectBanner = (props: { project: Project }) => {
    if (!props.project) return <div/>;
    const youtubeLink = props.project.YoutubeLink;
    const bannerLink = props.project.BannerLink;

    return <div className={"d-flex justify-content-center mb-2"}>
        {bannerLink == "" ?
            <div>
                <a href={youtubeLink} className="text-light text-center">
                    <h1 className="title">{props.project.Title}</h1>
                </a>
                <h3 className="text-center">{props.project.Sources}</h3>
            </div> :
            <a href={youtubeLink} className="text-light text-center">
                <img src={bannerLink} className="mx-auto img-fluid" alt="banner"/>
            </a>}
    </div>;
};
