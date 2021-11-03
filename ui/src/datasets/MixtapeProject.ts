import {GuildMember} from "./GuildMember";
import {fetchApi} from "./hooks";

export type MixtapeProject = {
    id: string;
    mixtape: string;
    Name: string;
    blurb: string;
    channel: string;
    hosts: string[];
    tags: string[];
}

export const resolveHostNicks = (members: GuildMember[], project: MixtapeProject) =>
    members.filter(m => project.hosts.includes(m.user.id)).map(m => m.nick);

export const saveMixtapeProjects = (projects: MixtapeProject[]) => {
    return fetchApi("/mixtape/projects", {
        method: "POST",
        body: JSON.stringify(projects),
    });
};

export const deleteMixtapeProjects = (projects: MixtapeProject[]) => {
    return fetchApi("/mixtape/projects", {
        method: "DELETE",
        body: JSON.stringify(projects.map(x => x.id)),
    });
};
