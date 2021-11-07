import {isEmpty} from "lodash/fp";
import {Link} from "react-router-dom";
import {Project} from "../../datasets";

export const ProjectBanner = (props: { project: Project | undefined }) => {
    if (!props.project) return <div/>;
    return <div className={"d-flex justify-content-center mb-2"}>
        {isEmpty(props.project.BannerLink) ?
            <div>
                <BannerLink project={props.project}>
                    <h1 className="title">{props.project.Title}</h1>
                </BannerLink>
                <h3 className="text-center">{props.project.Sources}</h3>
            </div> :
            <BannerLink project={props.project}>
                <img src={props.project.BannerLink} className="mx-auto img-fluid" alt="banner"/>
            </BannerLink>
        }
    </div>;
};

const BannerLink = (props: { project: Project, children: JSX.Element }) =>
    isEmpty(props.project.YoutubeLink) ?
        <Link
            className="text-light text-center"
            to={`/projects/${props.project.Name}`}>
            {props.children}
        </Link> :
        <a
            className="text-light text-center"
            href={props.project.YoutubeLink}>
            {props.children}
        </a>;
