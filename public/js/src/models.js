import React from 'react'

const _ = require('lodash')

const Endpoint = '/api/v1'

export const ApiResponseTypes = {
    Error: "error",
    Projects: "projects",
    Parts: "parts",
    Directors: "directors",
    Sessions: "sessions"
}

class ApiResponse {
    Status
    Type
    Projects
    Parts
    Directors
    Sessions

    get = (key, defaultValue) => _.get(this, key, defaultValue)
    static fromJSON = (obj) => {
        const resp = apiObjectFromJSON(obj, new ApiResponse())
        switch (resp.Type) {
            case ApiResponseTypes.Error:
                resp.Error = ErrorResponse.fromJSON(resp.Error)
                break;
            case ApiResponseTypes.Projects:
                resp.Projects = ProjectsResponse.fromJSON(resp.Projects)
                break;
            case ApiResponseTypes.Parts:
                resp.Parts = PartsResponse.fromJSON(resp.Parts)
                break;
            case ApiResponseTypes.Directors:
                resp.Directors = DirectorsResponse.fromJSON(resp.Directors)
                break;
            case ApiResponseTypes.Sessions:
                resp.Sessions = SessionsResponse.fromJSON(resp.Sessions)
                break;
        }
        console.log(resp)
        return resp
    }
}

const apiObjectFromJSON = (obj, dest) => {
    const cleanMap = _.keys(obj).reduce((a, b) => a.set(_.snakeCase(b), obj[b]), new Map())
    _.keys(dest).forEach(k => dest[k] = cleanMap.get(_.snakeCase(k)))
    return dest
}

class ErrorResponse {
    Code
    Error

    static fromJSON = (obj) => apiObjectFromJSON(obj, new ErrorResponse())
}

class ProjectsResponse {
    Projects = []

    static fromJSON = (obj) => {
        const resp = apiObjectFromJSON(obj, new ProjectsResponse())
        resp.Projects = _.get(resp, 'Projects', []).map(p => Project.fromJSON(p))
        return resp
    }
}

class Project {
    Name
    Title
    VideoReleased
    YoutubeLink
    YoutubeEmbed
    BannerLink
    Composers
    Arrangers
    PartsPage
    Sources
    PartsArchived
    PartsReleased
    Hidden

    static fromJSON = (obj) => apiObjectFromJSON(obj, new Project())

    isOpenForSubmission() {
        if (this.Hidden) return false
        if (this.VideoReleased) return false
        if (this.PartsArchived) return false
        return this.PartsReleased
    }
}


export const latestProject = (projects) => {
    if (projects) {
        const released = projects.filter(proj => proj.VideoReleased === true)
        released.sort()
        return released.pop()
    }
}

export class PartsResponse {
    Parts = []

    static fromJSON = (obj) => {
        const resp = apiObjectFromJSON(obj, new PartsResponse())
        resp.Parts = _.get(resp, 'Parts', []).map(p => Part.fromJSON(p))
        return resp
    }
}

export class Part {
    Project
    PartName
    ScoreOrder
    SheetMusicFile
    ClickTrackFile
    ConductorVideo
    PronunciationGuide
    SheetMusicLink
    ClickTrackLink
    PronunciationGuideLink

    static fromJSON = (obj) => apiObjectFromJSON(obj, new Part())
}

export class DirectorsResponse {
    Directors = []

    static fromJSON = (obj) => {
        const resp = apiObjectFromJSON(obj, new DirectorsResponse())
        resp.Directors = _.get(resp, 'Directors', []).map(p => Director.fromJSON(p))
        return resp
    }
}

export class Director {
    Name
    Epithet
    Affiliations
    Blurb
    Icon

    static fromJSON = (obj) => apiObjectFromJSON(obj, new Director())
}


export class SessionsResponse {
    Sessions = []

    static fromJSON = (obj) => {
        const resp = apiObjectFromJSON(obj, new SessionsResponse())
        resp.Sessions = _.get(resp, 'Sessions', []).map(p => Session.fromJSON(p))
        return resp
    }
}

const __sessionKinds = Object.freeze({
    Password: "password",
    Bearer: "bearer",
    Basic: "basic",
    Discord: "discord",
})

export const SessionKinds = __sessionKinds

export class Session {
    Key
    Kind
    Roles
    DiscordID
    Expires

    static fromJSON = (obj) => {
        const dest = apiObjectFromJSON(obj, new Session())
        if (dest.Expires) dest.Expires = new Date(dest.Expires)
        return dest
    }
}

export const deleteSessions = async (sessions) => {
    const payload = JSON.stringify(({'sessions': sessions}))
    return fetch(Endpoint + "/sessions", {
        method: 'DELETE',
        headers: {'Content-Type': 'application/json'},
        body: payload
    }).then(resp => resp.json()).then(data => {
        const response = ApiResponse.fromJSON(data)
        if (response.Type === ApiResponseTypes.Error) {
            throw 'vvgo.org error: ' + response.Error
        }
    })
}

export class SpreadsheetResponse {
    Spreadsheet
}

export class Credit {
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

export const useParts = () => _.get(useApiData(Endpoint + "/parts"), 'Parts', new PartsResponse())
export const useProjects = () => _.get(useApiData(Endpoint + "/projects"), 'Projects', new ProjectsResponse())
export const useDirectors = () => _.get(useApiData(Endpoint + "/leaders"), 'Directors', new DirectorsResponse())
export const useSessions = () => _.get(useApiData(Endpoint + "/sessions"), 'Sessions', new SessionsResponse())

export const useCredits = (project) => {
    const url = (project !== undefined && project.name !== undefined) ? `${Endpoint + "/credits"}?project=${project.name}` : undefined
    return useApiData(url)
}


const useApiData = (url) => {
    const [data, setData] = React.useState(new ApiResponse())
    React.useEffect(() => {
        fetch(url, {
            method: 'GET'
        }).then(response =>
            response.json()
        ).then(obj => {
            const response = ApiResponse.fromJSON(obj)
            if (response.Type === ApiResponseTypes.Error) {
                throw `vvgo.org error: [${response.Error.Code}] ${response.Error.Error}`
            }
            setData(response)
        })
    }, [url])
    return data
}

