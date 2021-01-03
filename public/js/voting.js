$(document).ready(() => {
    const createChoiceItem = (choice) => {
        const li = document.createElement("li")
        li.id = choice
        li.classList.add("vote")
        const title = document.createElement('div')
        title.classList.add('title', 'text-left')
        title.append(choice.split(' - ')[0])
        const subtext = document.createElement('small')
        subtext.append(choice.split(' - ')[1])
        li.append(title, subtext)
        return li
    }

    fetch('/api/v1/arrangements/ballot')
        .then(resp => resp.json())
        .then(data => {
            const element = document.createElement("ol")
            element.id = "sortable"
            element.append(...data.map(choice => createChoiceItem(choice)))
            document.getElementById("ballot").append(element)

            let sortable = $("#sortable")
            sortable.sortable({cursor: "ns-resize", axis: "y"})
            sortable.disableSelection()

            hideElement('loading')
        })

    showLoadingText('loading')
})

function submitVote() {
    hideElement('voteSubmitted')
    hideElement('submissionFailed')

    const ballot = [...document.getElementsByClassName("vote")].map(vote => vote.id)
    fetch('/api/v1/arrangements/ballot', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(ballot)
    }).then((response) => {
        if (response.status !== 200) {
            console.log("invalid ballot", ballot)
            showElement('submissionFailed')
        } else {
            console.log("ballot submitted", ballot)
            showElement('voteSubmitted')
        }
    }).catch(error => {
        console.log(error)
        showElement('submissionFailed')
    })
}

function showLoadingText(elementId) {
    const choices = [
        "😩 【Ｌｏａｄｉｎｇ】 😩", "(っ◔◡◔)っ ♥ 𝐿𝑜𝒶𝒹𝒾𝓃𝑔 ♥", "𝒲𝑒'𝓁𝓁 𝒷𝑒 𝓇𝒾𝑔𝒽𝓉 𝓌𝒾𝓉𝒽 𝓎𝑜𝓊 😘",
        "😳👌  ⓛＯα𝓓𝕚ＮＧ  💗🍩", "🐏  🎀  𝒯𝐻𝐸 𝐸𝒜𝑅𝒯𝐻 𝐼𝒮 𝐹𝐿𝒜𝒯  🎀  🐏"
    ]
    document.getElementById(elementId).append(choices[Math.floor(Math.random() * choices.length)])
}

function showElement(elementId) {
    document.getElementById(elementId).classList.remove('d-none')
}

function hideElement(elementId) {
    document.getElementById(elementId).classList.add('d-none')
}
