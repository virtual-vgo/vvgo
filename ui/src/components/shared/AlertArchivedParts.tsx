import {Project} from "../../datasets";

export const AlertArchivedParts = (props: { project: Project }) => props.project.PartsArchived ?
    <div className="alert alert-warning">
        This project has been archived. Parts are only visible to executive directors.
    </div> : <div/>;
