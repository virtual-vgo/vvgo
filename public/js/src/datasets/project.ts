export class Project {
    Name: string
    Title: string
    VideoReleased: boolean
    YoutubeLink: string
    YoutubeEmbed: string
    BannerLink: string
    Composers: string
    Arrangers: string
    PartsPage: string
    Sources: string
    PartsArchived: boolean
    PartsReleased: boolean
    Hidden: boolean
}

export const projectIsOpenForSubmission = (project: Project): boolean => {
    if (project.Hidden) return false
    if (project.VideoReleased) return false
    if (project.PartsArchived) return false
    return project.PartsReleased
}
export const projectIsPostProduction = (project: Project): boolean => {
    if (project.Hidden) return false
    if (project.VideoReleased) return false
    return project.PartsArchived
}
export const projectIsReleased = (project: Project): boolean => {
    if (project.Hidden) return false
    return project.VideoReleased
}

export const latestProject = (projects: Project[]): Project => {
    if (projects) {
        const released = projects.filter(proj => proj.VideoReleased === true)
        released.sort()
        return released.pop()
    }
}

