import {Project} from "../../datasets";

export const YoutubeIframe = (props: { project: Project }) => {
    const latest = props.project;
    if (!latest) return <div/>;
    if (latest.YoutubeEmbed === "") return <div/>;
    return <div className="project-iframe-wrapper text-center m-2">
        <iframe className="project-iframe" src={latest.YoutubeEmbed}
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                allowFullScreen/>
    </div>;
};
