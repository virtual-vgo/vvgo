import { Link } from "react-router-dom";
import { Channels } from "../../data/discord";
import { LinkChannel } from "../shared/LinkChannel";

export const LoginFailure = () => {
  return (
    <p>
      Please join our <a href="https://discord.gg/vvgo">Discord server</a> and
      accept the rules before logging in with Discord.
      <br />
      If you think you should be able to login, please check{" "}
      <LinkChannel channel={Channels.GeekSquad} />.
      <br />
      <br />
      <Link to="/login">Return to the login page.</Link>
    </p>
  );
};

export default LoginFailure;
