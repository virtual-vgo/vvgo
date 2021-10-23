import {fetchApi} from "./hooks";

export interface MixtapeProject {
    Name: string;
    Blurb: string;
    Channel: string;
    Owners: string[];
    Tags: string[];
}

export const saveMixtapeProject = (project: MixtapeProject) => {
    return fetchApi("/mixtape", {
        method: "POST",
        body: JSON.stringify([project]),
    });
};
