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
        p.append(entry['name'])
        p.classList.add("text-light")

        if (entry['discord_id'] != null && entry['discord_id'] !== "") {
            p.append(document.createElement("br"))
            p.append(createDeleteLink(entry))
        }

        const td = document.createElement("td")
        td.append(p)
        return td
    }

    const createTitle = (entry) => {
        const p = document.createElement("p")
        p.append(entry['title'])
        p.classList.add("text-light")

        const td = document.createElement("td")
        td.append(p)
        return td
    }

    const createBlurb = (entry) => {
        const p = document.createElement("p")
        p.append(entry['blurb'])
        const td = document.createElement("td")
        td.append(p)
        return td
    }

    const createDeleteLink = (entry) => {
        const a = document.createElement("a")
        a.classList.add("text-light", "text-sm")
        a.append("(delete)")
        a.addEventListener("click", (e) => {
            fetch('/api/v1/aboutme', {method: 'DELETE', body: JSON.stringify([entry['discord_id']])})
                .then(createAboutMe)
        })
        a.style.cursor = 'pointer'
        return a
    }

    const createAboutmeRow = (entry, isFirst, isLast) => {
        const tr = document.createElement("tr")
        if (isFirst === false) tr.classList.add("border-top")
        if (isLast === false) tr.classList.add("border-bottom")
        tr.append(createName(entry), createTitle(entry), createBlurb(entry))
        return tr
    }

    const createAboutmeTbody = (entries) => {
        const table = document.getElementById("aboutme-table")
        while (table.firstChild) {
            table.removeChild(table.firstChild);
        }

        if (entries == null) {
            return
        }
        const tbody = document.createElement("tbody")
        tbody.append(...entries.map(
            entry => createAboutmeRow(entry, entry === entries[0], entry === entries[entries.length - 1])
        ))

        table.append(tbody)
    }

    const createAboutMe = () => {
        fetch('/api/v1/roles')
            .then(resp => resp.json())
            .then(roles => {
                if (roles.includes("vvgo-leader")) {
                    fetch('/api/v1/aboutme')
                        .then(resp => resp.json())
                        .then(data => createAboutmeTbody(data))
                }
            })
    }

    createAboutMe()
}

