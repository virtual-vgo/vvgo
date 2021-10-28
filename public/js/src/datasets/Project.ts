export interface Project {
    Name: string;
    Title: string;
    Season: string;
    Hidden: boolean;
    PartsReleased: boolean;
    PartsArchived: boolean;
    VideoReleased: boolean;
    Sources: string;
    Composers: string;
    Arrangers: string;
    Editors: string;
    Transcribers: string;
    Preparers: string;
    ClixBy: string;
    Reviewers: string;
    Lyricists: string;
    AdditionalContent: string;
    ReferenceTrack: string;
    ChoirPronunciationGuide: string;
    BannerLink: string;
    YoutubeLink: string;
    YoutubeEmbed: string;
    SubmissionDeadline: string;
    SubmissionLink: string;
    ReferenceTrackLink: string;
}

export const projectIsOpenForSubmission = (project: Project): boolean => {
    if (project.Hidden) return false;
    if (project.VideoReleased) return false;
    if (project.PartsArchived) return false;
    return project.PartsReleased;
};

export const projectIsPostProduction = (project: Project): boolean => {
    if (project.Hidden) return false;
    if (project.VideoReleased) return false;
    return project.PartsArchived;
};

export const projectIsReleased = (project: Project): boolean => {
    if (project.Hidden) return false;
    return project.VideoReleased;
};

export const latestProject = (projects: Project[]): Project => {
    if (projects) {
        const released = projects.filter(proj => proj.VideoReleased === true);
        released.sort();
        return released.pop();
    }
    return null;
};
