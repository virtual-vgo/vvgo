import React from 'react'

const Endpoint = '/api/v1'

const StatusOk = "ok"
const StatusError = "error"
const ResponseTypeError = "error"
const ResponseTypeSessions = "sessions"

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

export const useProjects = () => useApiData(Endpoint + "/projects", Project)

export const latestProject = (projects) => {
    console.log(projects)
    const released = projects.filter(proj => proj.videoReleased === true)
    released.sort()
    return released.pop()
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

export const useCredits = (project) => {
    const url = (project !== undefined && project.name !== undefined) ? `${Endpoint + "/credits"}?project=${project.name}` : undefined
    return useApiData(url, TopicCreditsRow)
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

class ExecutiveDirector {
    constructor(obj) {
        this.Name = obj['Name']
        this.Epithet = obj['Epithet']
        this.Affiliations = obj['Affiliations']
        this.Blurb = obj['Blurb']
        this.Icon = obj['Icon']
    }
}

export const useDirectors = () =>
    useApiData(Endpoint + "/leaders", ExecutiveDirector)

class Session {
    constructor(obj) {
        this.Key = obj['key']
        this.Kind = obj['kind']
        this.Roles = obj['roles']
        this.DiscordID = obj['discord_id']
    }
}

export const useSessions = () =>
    useApiData(Endpoint + "/sessions", Session)

export const deleteSessions = async (sessions) => {
    const payload = JSON.stringify(({'sessions': sessions}))
    return fetch(Endpoint + "/sessions", {
        method: 'DELETE',
        headers: {'Content-Type': 'application/json'},
        body: payload
    }).then(resp => resp.json()).then(data => {
        const response = new Response(data)
        if (response.Type === ResponseTypeError) {
            throw 'vvgo.org error: ' + response.Error
        }
    })
}

const useApiData = (url, decoder) => {
    const [data, setData] = React.useState([])
    React.useEffect(() => {
        fetch(url)
            .then(response => response.json())
            .then(jsonData => jsonData.map(obj => new decoder(obj)))
            .then(decoded => setData(decoded))
            .catch(error => console.log(error))
    }, [url, decoder])
    return data
}

class Response {
    constructor(obj) {
        this.Status = obj['status']
        this.Type = obj['type']

        switch (this.Type) {
            case ResponseTypeError:
                this.Error = new ErrorResponse(obj['error'])
                break
            case ResponseTypeSessions:
                this.Sessions = new SessionResponse(obj['sessions'])
                break
        }
    }
}

class ErrorResponse {
    constructor(obj) {
        this.Code = obj['code']
        this.Error = obj['error']
    }
}

class SessionResponse {
    constructor(obj) {
        this.Status = obj['status']
        this.Type = obj['type']
        this.Error = obj['error']
        this.Session = obj['session']
    }
}
