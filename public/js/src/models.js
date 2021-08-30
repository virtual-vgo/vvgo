import React from 'react'

const ProjectsEndpoint = '/api/v1/projects'
const CreditsEndpoint = '/api/v1/credits'

class Project {
    constructor(obj) {
        this.name = obj['Name']
        this.title = obj['Title']
        this.videoReleased = obj['VideoReleased']
        this.youtubeLink = obj['YoutubeLink']
        this.youtubeEmbed = obj['YoutubeEmbed']
        this.bannerLink = obj['BannerLink']
        this.composers = obj['Composers']
        this.arrangers = obj['Arrangers']
        this.partsPage = obj['PartsPage']
        this.sources = obj['Sources']
        this.partsArchived = obj['PartsArchived']
        this.partsReleased = obj['PartsReleased']
        this.hidden = obj['Hidden']
    }

    isOpenForSubmission() {
        if (this.hidden) return false
        if (this.videoReleased) return false
        if (this.partsArchived) return false
        return this.partsReleased
    }
}

export const latestProject = (projects) => {
    console.log(projects)
    const released = projects.filter(proj => proj.videoReleased === true)
    released.sort()
    return released.pop()
}

export const useProjects = () => {
    const [data, setData] = React.useState([])
    React.useEffect(() => {
        fetch(ProjectsEndpoint)
            .then(response => response.json())
            .then(data => data.map(obj => new Project(obj)))
            .then(projects => setData(projects))
            .catch(error => console.log(error))
    }, [ProjectsEndpoint])
    return [data, setData]
}

class Credit {
    constructor(obj) {
        this.project = obj['Project']
        this.order = obj['Order']
        this.name = obj['Name']
        this.majorCategory = obj['MajorCategory']
        this.minorCategory = obj['MinorCategory']
        this.bottomText = obj['BottomText']
    }
}

class TeamCreditsRow {
    constructor(obj) {
        this.name = obj['Name']
        this.rows = obj['Rows'].map(credit => new Credit(credit))
    }
}

class TopicCreditsRow {
    constructor(obj) {
        this.name = obj['Name']
        this.rows = obj['Rows'].map(teamRow => new TeamCreditsRow(teamRow))
    }
}

export const useCredits = (project) => {
    const [data, setData] = React.useState([])
    const url = (project !== undefined && project.name !== undefined) ? `${CreditsEndpoint}?project=${project.name}` : undefined
    React.useEffect(() => {
        if (url !== undefined)
            fetch(url)
                .then(response => response.json())
                .then(data => data.map(obj => new TopicCreditsRow(obj)))
                .then(credits => setData(credits))
                .catch(error => console.log(error))
    }, [url])
    return [data, setData]
}
