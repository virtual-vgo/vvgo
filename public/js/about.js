$(document).ready(() => {
    createLeadersTbody()
    createAboutmeTBody()
})

function createLeadersTbody() {
    const createIcon = (leader) => {
        const img = document.createElement("img")
        img.src = leader['Icon']
        img.alt = leader['Name']
        img.height = 100

        const td = document.createElement("td")
        td.append(img)
        return td
    }

    const createName = (leader) => {
        const br = document.createElement("br")
        const small = document.createElement("small")
        small.append(leader['Epithet'])

        const p = document.createElement("p")
        p.append(leader['Name'], br, small)
        p.classList.add("text-light")

        const td = document.createElement("td")
        td.append(p)
        return td
    }

    const createBlurb = (leader) => {
        const blurb = (() => {
            const p = document.createElement("p")
            p.append(leader['Blurb'])
            return p
        })()

        const affiliations = (() => {
            const i = document.createElement("i")
            i.append(leader['Affiliations'])
            const p = document.createElement("p")
            p.append(i)
            return p
        })()

        const td = document.createElement("td")
        td.append(blurb, affiliations)
        return td
    }

    const createLeaderRow = (leader, isFirst, isLast) => {
        const tr = document.createElement("tr")
        if (isFirst === false) tr.classList.add("border-top")
        if (isLast === false) tr.classList.add("border-bottom")
        tr.append(createIcon(leader), createName(leader), createBlurb(leader))
        return tr
    }

    fetch('/api/v1/leaders')
        .then(resp => resp.json())
        .then(data => {
            const element = document.createElement("tbody")
            element.append(...data.map(
                leader => createLeaderRow(leader, leader === data[0], leader === data[data.length - 1])
            ))
            document.getElementById("leader-table").append(element)
        })
}

function createAboutmeTBody() {
    const createName = (entry) => {
        const p = document.createElement("p")
        p.append(entry['Name'])
        p.classList.add("text-light")

        const td = document.createElement("td")
        td.append(p)
        return td
    }

    const createBlurb = (entry) => {
        const p = document.createElement("p")
        p.append(entry['Blurb'])

        const td = document.createElement("td")
        td.append(p)
        return td
    }

    const createAboutmeRow = (entry, isFirst, isLast) => {
        const tr = document.createElement("tr")
        if (isFirst === false) tr.classList.add("border-top")
        if (isLast === false) tr.classList.add("border-bottom")
        tr.append(createName(entry), createBlurb(entry))
        return tr
    }

    fetch('/api/v1/aboutme')
        .then(resp => resp.json())
        .then(data => {
            const element = document.createElement("tbody")
            element.append(...data.map(
                entry => createAboutmeRow(entry, entry === data[0], entry === data[data.length - 1])
            ))
            document.getElementById("aboutme-table").append(element)
        })
}
