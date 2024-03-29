import Col from "react-bootstrap/Col";
import Row from "react-bootstrap/Row";
import { Director, useDirectors } from "../datasets";
import { LoadingText } from "./shared/LoadingText";

export const About = () => {
  const directors = useDirectors();
  return (
    <div>
      <Blurb />
      <Directors directors={directors} />
    </div>
  );
};

const Blurb = () => {
  return (
    <Row className="border-light row-cols-2">
      <Col lg={3}>
        <h2 className="text-center">About Us</h2>
      </Col>
      <div className="col col-lg-9 col-12 text-justify">
        <p className="">
          Formed in March 2020, <strong>Virtual Video Game Orchestra</strong>{" "}
          (VVGO, &quot;vee-vee-go&quot;) is an online volunteer-run music
          ensemble predicated on providing a musical performance outlet for
          musicians whose IRL rehearsals and performances were cancelled due to
          COVID-19. Led and organized by members from various video game
          ensembles, and with a community of hundreds of musicians from across
          the globe, VVGO is open to any who wish to participate regardless of
          instrument, skill level, or musical background.
        </p>
        <p className="">
          Our mission is to provide a fun and accessible virtual community of
          musicians from around the world through performing video game music.
        </p>
        <p className="">
          We are always accepting new members into our community. If you would
          like to join our orchestra or get more information about our current
          performance opportunities, please join us on{" "}
          <a href="https://discord.gg/mt8EZdsrc6" className="text-info">
            Discord
          </a>
          !
        </p>
      </div>
    </Row>
  );
};

const Directors = (props: { directors?: Director[] }) => {
  return (
    <div className="text-center">
      <Row className="mt-3 row-cols-2">
        <Col lg={3}>
          <h2>VVGO Leadership</h2>
        </Col>
        <Col lg={9} md={12}>
          {props.directors ? (
            <ExecutiveDirectorTable directors={props.directors} />
          ) : (
            <LoadingText />
          )}
        </Col>
      </Row>
    </div>
  );
};

const ExecutiveDirectorTable = (props: { directors: Director[] }) => {
  return (
    <table
      id="leader-table"
      className="table table-responsive table-borderless text-light"
    >
      <tbody>
        {props.directors.map((director, i) => (
          <ExecutiveDirectorRow
            key={director.name}
            director={director}
            bottom={props.directors.length == i + 1}
          />
        ))}
      </tbody>
    </table>
  );
};

const ExecutiveDirectorRow = (props: {
  director: Director;
  bottom: boolean;
}) => {
  return (
    <tr className={props.bottom ? "" : "border-bottom"}>
      <td>
        <img src={props.director.icon} alt={props.director.name} height="100" />
      </td>
      <td>
        <p className="text-light">
          {props.director.name}
          <br />
          <small>{props.director.epithet}</small>
        </p>
      </td>
      <td>
        <p>{props.director.blurb}</p>
        <p>
          <i>{props.director.affiliations}</i>
        </p>
      </td>
    </tr>
  );
};

export default About;
