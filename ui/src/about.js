import React from 'react'

const axios = require('axios').default;

class About extends React.Component {
    constructor(props) {
        super(props)
        this.state = {leaders: []}
    }

    componentDidMount() {
        axios.get('/leaders').then(response => this.setState({leaders: response.data}))
    }

    InfoRow() {
        return <div className="row mt-4 text-justify">
            <div className="col">
                <p className="blockquote">
                    Formed in March 2020, <strong>Virtual Video Game Orchestra</strong> (VVGO, "vee-vee-go") is an
                    online
                    volunteer-run music ensemble predicated on providing a musical performance outlet for musicians
                    whose
                    IRL rehearsals and performances were cancelled due to COVID-19. Led and organized by members
                    from
                    various video game ensembles, and with a community of hundreds of musicians from across the
                    globe,
                    VVGO is open to any who wish to participate regardless of instrument, skill level, or musical
                    background.
                </p>
                <p className="blockquote">
                    Our mission is to provide a fun and accessible virtual community of musicians from around the
                    world
                    through performing video game music.
                </p>
                <p className="blockquote">
                    We are always accepting new members into our community. If you would like to join our orchestra
                    or
                    get more information about our current performace opportunities, please join us on
                    <a href="https://discord.gg/9RVUJMQ" className="text-info">Discord</a>!
                </p>
            </div>
        </div>
    }


    LeaderRow(leader) {
        let nameData = <p className="text-light">
            {leader.Name}<br/><small>{leader.Epithet}</small>
        </p>

        if (leader.Email !== "") {
            let href = "mailto: " + leader.Email
            nameData = <a className="text-light" href={href}>
                {leader.Name}<br/><small>{leader.Epithet}</small>
            </a>
        }

        return <tr key={leader.Name}>
            <td><img src={leader.Icon} alt={leader.Name} height="125"/></td>
            <td>
                {nameData}
            </td>
            <td>
                <p>{leader.Blurb}</p>
                <p><i>{leader.Affiliations}</i></p>
            </td>
        </tr>
    }

    render() {
        console.log(this.state.leaders)
        let leaderRows = this.state.leaders.map(leader => this.LeaderRow(leader))
        return <div className="container">
            {this.InfoRow()}
            <div className="row">
                <div className="col text-center"><h2>VVGO Leadership</h2></div>
            </div>
            <div className="row justify-content-md-center">
                <div className="col col-md-auto mt-4 text-center">
                    <table className="table table-bordered table-responsive text-light">
                        <tbody>
                        {leaderRows}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    }
}

export default About

