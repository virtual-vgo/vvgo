$(document).ready(() => {
    const createIcon = (leader) => {
        const td = document.createElement("td")
        const img = document.createElement("img")
        img.src = leader['Icon']
        img.alt = leader['Name']
        img.height = 100
        td.append(img)
        return td
    }

    const createName = (leader) => {
        const td = document.createElement("td")
        const p = document.createElement("p")
        const br = document.createElement("br")
        const small = document.createElement("small")
        small.append(leader['Epithet'])
        p.append(leader['Name'], br, small)
        p.classList.add("text-light")
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
        if (!isFirst) {
            tr.classList.add("border-top")
        }
        if (!isLast)
            tr.classList.add("border-bottom")
        tr.append(createIcon(leader), createName(leader), createBlurb(leader))
        return tr
    }

    fetch('/api/v1/leaders')
        .then(resp => resp.json())
        .then(data => {
            const element = document.createElement("tbody")
            element.append(...data.map(leader => createLeaderRow(
                leader, leader === data[0], leader === data[data.length - 1])))
            document.getElementById("leader-table").append(element)
        })
})
