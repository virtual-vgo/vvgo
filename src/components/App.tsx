import isEmpty from "lodash/fp/isEmpty";
import { lazy, Suspense } from "react";
import { BrowserRouter, Redirect, Route, Switch } from "react-router-dom";
import { getSession, updateLogin } from "../auth";
import { UserRole } from "../datasets";
import Admin from "./admin/Admin";
import Contact from "./Contact";
import AccessDenied from "./errors/AccessDenied";
import InternalOopsie from "./errors/InternalOopsie";
import NotFound from "./errors/NotFound";
import Home from "./Home";
import Login from "./login/Login";
import LoginDiscord from "./login/LoginDiscord";
import LoginFailure from "./login/LoginFailure";
import Logout from "./login/Logout";
import { Footer } from "./shared/Footer";
import { LoadingText } from "./shared/LoadingText";
import { Navbar } from "./shared/Navbar";

// Lazy loads
const MemberStats = lazy(() => import("./stats/Members"));
const Projects = lazy(() => import("./Projects"));
const Parts = lazy(() => import("./Parts"));
const NewProjectWorkflow = lazy(() => import("./mixtape/NewProjectWorkflow"));
const MemberDashboard = lazy(() => import("./mixtape/MemberDashboard"));
const About = lazy(() => import("./About"));
const CreditsMaker = lazy(() => import("./CreditsMaker"));
const Sessions = lazy(() => import("./admin/Sessions"));
const ManageMixtapes = lazy(() => import("./admin/Mixtapes"));

export const App = () => {
  updateLogin();
  return (
    <BrowserRouter>
      <Switch>
        <PrivateRoute path="/admin" role={UserRole.ExecutiveDirector}>
          <AdminRoutes />
        </PrivateRoute>

        <PrivateRoute path="/credits-maker/" role={UserRole.ProductionTeam}>
          <AppPage title="Credits Maker">
            <CreditsMaker />
          </AppPage>
        </PrivateRoute>

        <PrivateRoute path="/stats/members" role={UserRole.VerifiedMember}>
          <AppPage title="Member Stats">
            <MemberStats />
          </AppPage>
        </PrivateRoute>

        <PrivateRoute path="/parts/" role={UserRole.VerifiedMember}>
          <AppPage title="Parts">
            <Parts />
          </AppPage>
        </PrivateRoute>

        <PrivateRoute path="/mixtape/" role={UserRole.VerifiedMember}>
          <MixtapeRoutes />
        </PrivateRoute>

        <Route path="/projects/">
          <AppPage title="Projects">
            <Projects />
          </AppPage>
        </Route>

        <Route path="/about/">
          <AppPage title="About">
            <About />
          </AppPage>
        </Route>

        <Route path="/contact/">
          <AppPage title="Contact">
            <Contact />
          </AppPage>
        </Route>

        <Route path="/logout/">
          <AppPage title="Logout">
            <Logout />
          </AppPage>
        </Route>

        <Route path="/login/">
          <LoginRoutes />
        </Route>

        <Route exact path="/">
          <AppPage>
            <Home />
          </AppPage>
        </Route>
        <Route exact path="/401.html">
          <AccessDenied />
        </Route>
        <Route exact path="/404.html">
          <NotFound />
        </Route>
        <Route exact path="/500.html">
          <InternalOopsie />
        </Route>
        <Route path="*">
          <NotFound />
        </Route>
      </Switch>
    </BrowserRouter>
  );
};

const AdminRoutes = () => {
  return (
    <Switch>
      <PrivateRoute path="/admin/mixtape/" role={UserRole.ExecutiveDirector}>
        <AppPage title="Manage Mixtape Projects">
          <ManageMixtapes />
        </AppPage>
      </PrivateRoute>

      <PrivateRoute path="/admin/sessions/" role={UserRole.ExecutiveDirector}>
        <AppPage title="Sessions">
          <Sessions />
        </AppPage>
      </PrivateRoute>

      <Route path="/admin/">
        <AppPage title="Admin Links">
          <Admin />
        </AppPage>
      </Route>
    </Switch>
  );
};

const MixtapeRoutes = () => {
  return (
    <Switch>
      <PrivateRoute
        path="/mixtape/NewProjectWorkflow/"
        role={UserRole.ExecutiveDirector}
      >
        <AppPage title="New Project Workflow">
          <NewProjectWorkflow />
        </AppPage>
      </PrivateRoute>
      <Route>
        <AppPage title="Mixtape">
          <MemberDashboard />
        </AppPage>
      </Route>
    </Switch>
  );
};

const LoginRoutes = () => {
  return (
    <Switch>
      <Route path="/login/failure/">
        <LoginFailure />
      </Route>
      <Route path="/login/discord/">
        <LoginDiscord />
      </Route>
      <Route>
        <AppPage title="Login">
          <Login />
        </AppPage>
      </Route>
      <Route path="*">
        <NotFound />
      </Route>
    </Switch>
  );
};

const PrivateRoute = (props: {
  path: string;
  role: string;
  children: JSX.Element;
}) => {
  const me = getSession();
  const authorized = me.roles?.includes(props.role) ?? false;
  const children = () => {
    switch (true) {
      case authorized:
        return props.children;
      case me.isAnonymous():
        return <Redirect to={"/login"} />;
      default:
        return <AccessDenied />;
    }
  };
  return <Route path={props.path} render={children} />;
};

function AppPage(props: { title?: string; children: JSX.Element }) {
  const title = isEmpty(props.title)
    ? "Virtual Video Game Orchestra"
    : props.title;
  document.title = "VVGO | " + title;
  return (
    <div className={"container"}>
      <Navbar />
      <Suspense fallback={<LoadingText />}>{props.children}</Suspense>
      <Footer />
    </div>
  );
}
