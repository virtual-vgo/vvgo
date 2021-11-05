import _ = require("lodash");

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

export const latestProject = (projects: Project[]): Project =>
    <Project>_.defaultTo(projects, [])
        .filter(proj => proj.VideoReleased)
        .sort((a, b) => a.Name.localeCompare(b.Name))
        .pop();
