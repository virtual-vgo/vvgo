import React from "react";

const axios = require('axios').default;

function Banner(props) {
    return <a href={props.YoutubeLink}>
        <img src={props.BannerLink} className="mx-auto img-fluid" alt="banner"/>
    </a>;
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

export {Banner, YoutubeIframe, GetRoles}
