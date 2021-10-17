import React from 'react'

const _ = require('lodash')

const Endpoint = '/api/v1'

export const ApiResponseStatus = Object.freeze({
    OK: "ok",
    Error: "error",
})

class ApiResponse {
    Status
    Projects
    Parts
    Directors
    Sessions
    Identity

    static fromJSON = (obj) => {
        const resp = apiObjectFromJSON(obj, new ApiResponse())
        if (resp.Status === ApiResponseStatus.Error) {
            resp.Error = ErrorResponse.fromJSON(_.get(resp, 'Error', {'error': 'unknown'}))
        } else {
            resp.Projects = _.get(resp, 'Projects', []).map(p => Project.fromJSON(p))
            resp.Parts = _.get(resp, 'Parts', []).map(p => Part.fromJSON(p))
            resp.Directors = _.get(resp, 'Directors', []).map(p => Director.fromJSON(p))
            resp.Sessions = _.get(resp, 'Sessions', []).map(p => Session.fromJSON(p))
            resp.Identity = Session.fromJSON(_.get(resp, 'Identity', {}))
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

export class Director {
    Name
    Epithet
    Affiliations
    Blurb
    Icon

    static fromJSON = (obj) => apiObjectFromJSON(obj, new Director())
}

export const SessionKinds = Object.freeze({
    Password: "password",
    Bearer: "bearer",
    Basic: "basic",
    Discord: "discord",
    ApiToken: "api_token"
})

export class Session {
    Key
    Kind
    Roles
    DiscordID
    ExpiresAt

    static fromJSON = (obj) => {
        const dest = apiObjectFromJSON(obj, new Session())
        if (dest.ExpiresAt) dest.ExpiresAt = new Date(dest.ExpiresAt)
        return dest
    }
}

export const createSessions = async (sessions) => {
    const payload = JSON.stringify(({'sessions': sessions}))
    return fetch(Endpoint + "/sessions", {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: payload
    }).then(resp => resp.json()).then(data => {
        const response = ApiResponse.fromJSON(data)
        if (response.Type === ApiResponseStatus.Error) {
            throw 'vvgo.org error: ' + response.Error
        }
        return response.Sessions
    })
}

export const deleteSessions = async (sessionsId) => {
    const payload = JSON.stringify(({'sessions': sessionsId}))
    return fetch(Endpoint + "/sessions", {
        method: 'DELETE',
        headers: {'Content-Type': 'application/json'},
        body: payload
    }).then(resp => resp.json()).then(data => {
        const response = ApiResponse.fromJSON(data)
        if (response.Type === ApiResponseStatus.Error) {
            throw 'vvgo.org error: ' + response.Error
        }
    })
}

export const useParts = () => useApiState(Endpoint + "/parts", 'Parts', [])
export const useProjects = () => useApiState(Endpoint + "/projects", 'Projects', [])
export const useDirectors = () => useApiState(Endpoint + "/leaders", 'Directors', [])
export const useSessions = () => useApiState(Endpoint + "/sessions", 'Sessions', [])
export const useMySession = () => useApiState(Endpoint + "/me", 'Identity', new Session())

export const useApiState = (url, key, defaultValue) => {
    const [data, setData] = useApiData(url)
    const setDataKey = (value) => {
        setData(_.set(data, key, value))
    }
    return [_.get(data, key, defaultValue), setDataKey]
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
            if (response.Type === ApiResponseStatus.Error) {
                throw `vvgo.org error: [${response.Error.Code}] ${response.Error.Error}`
            }
            setData(response)
        })
    }, [url])
    return [data, setData]
}

