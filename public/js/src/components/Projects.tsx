import React = require("react");
import {useProjects} from "../datasets";
import {LoadingText} from "./shared/LoadingText";
import {RootContainer} from "./shared/RootContainer";

export const Projects = () => {
    const projects = useProjects();

    return <RootContainer>
        <div>
            {projects ? projects.map(p => <p>{p.Title}</p>) : <LoadingText/>}
        </div>
    </RootContainer>;
};
