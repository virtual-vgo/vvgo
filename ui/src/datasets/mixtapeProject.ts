import _ from "lodash";
import {GuildMember} from "./GuildMember";
import {fetchApi} from "./hooks";

export interface mixtapeProject {
    Name: string;
    mixtape: string;
    title?: string;
    blurb?: string;
    channel?: string;
    hosts?: string[];
}

export const resolveHostNicks = (members: GuildMember[], project: mixtapeProject) =>
    _.uniq(members.filter(m => _.defaultTo(project.hosts, []).includes(m.user.id)).map(m => m.nick));

export const saveMixtapeProjects = (projects: mixtapeProject[]) => {
    return fetchApi("/mixtape/projects", {
        method: "POST",
        body: JSON.stringify(projects),
    });
};

export const deleteMixtapeProjects = (projects: mixtapeProject[]) => {
    return fetchApi("/mixtape/projects", {
        method: "DELETE",
        body: JSON.stringify(projects.map(x => x.Name)),
    });
};
