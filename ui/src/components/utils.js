import React from "react";

const axios = require('axios').default;

function ProjectBanner(props) {
    if (props.project.YoutubeLink !== "") {
        return <a href={props.project.YoutubeLink}>
            <img src={props.project.BannerLink} className="mx-auto img-fluid" alt="banner"/>
        </a>
    } else {
        return <div>
            <h2 className="title">{props.project.Title}</h2>
            <h3>{props.project.Sources}</h3>
        </div>
    }

}

function YoutubeIframe(props) {
    return <div className="project-iframe-wrapper text-center m-2">
        <iframe className="project-iframe" src={props.YoutubeEmbed}
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                title="latest.Title"
                allowFullScreen/>
    </div>;
}

function GetRoles() {
    const params = new URLSearchParams(window.location.search)
    const paramRoles = params.getAll("roles")
    return axios.post('/roles', paramRoles)
}

export {ProjectBanner, YoutubeIframe, GetRoles}
