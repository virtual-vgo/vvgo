import React = require("react");
import {Project} from "../../datasets";

export const AlertUnreleasedProject = (props: { project: Project }) => props.project.PartsReleased == false ?
    <div className="alert alert-warning">
        This project is unreleased and invisible to members!
    </div> : <div/>;
