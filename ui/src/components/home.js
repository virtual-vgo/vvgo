import React from "react";
import {ProjectBanner, YoutubeIframe} from "./utils";
import Container from "@material-ui/core/Container";

export default function Home(props) {
    let latestProject = {}
    props.projects.sort((a, b) => a.Name.localeCompare(b.Name))
        .forEach(project => {
            if (project.YoutubeEmbed !== "") {
                latestProject = project
            }
        })

    return <Container>
        <div className="row row-cols-1 justify-content-md-center text-center m-2">
            <div className="col">
                <ProjectBanner project={latestProject}/>
            </div>
            <div className="col">
                <YoutubeIframe YoutubeEmbed={latestProject.YoutubeEmbed}/>
            </div>
        </div>
        <div className="row justify-content-md-center text-center m-2">
            <div className="col text-center mt-2">
                <p>
                    If you would like to join our orchestra or get more information about our current projects,
                    please join us on <a href="https://discord.gg/9RVUJMQ">Discord!</a>
                </p>
            </div>
        </div>
    </Container>
}
