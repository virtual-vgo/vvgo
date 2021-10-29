import {fetchApi} from "./hooks";

export type MixtapeProject = {
    Id: string;
    Mixtape: string;
    Name: string;
    Blurb: string;
    Channel: string;
    Owners: string[];
    Tags: string[];
}

export const saveMixtapeProjects = (projects: MixtapeProject[]) => {
    return fetchApi("/mixtape", {
        method: "POST",
        body: JSON.stringify(projects),
    });
};

export const deleteMixtapeProjects = (projects: MixtapeProject[]) => {
    return fetchApi("/mixtape", {
        method: "DELETE",
        body: JSON.stringify(projects.map(x => x.Id)),
    });
};
