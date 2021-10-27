import React = require("react");
import {useParts, useProjects} from "../datasets";
import {RootContainer} from "./components";

export const Parts = () => {
    const projects = useProjects();
    const parts = useParts();

    return <RootContainer>
        {projects.map(r => <div>
            {r.Title}
            <table>
                <tbody>
                {parts.map(r => <tr>
                    <td>{r.PartName}</td>
                    <td>{r.Project}</td>
                </tr>)}
                </tbody>
            </table>
        </div>)}
    </RootContainer>;
};
