name := "isuketch"
organization := "net.isucon.isucon6"
version := "1.0"
scalaVersion := "2.11.8"

scalacOptions := Seq("-unchecked", "-deprecation", "-encoding", "utf8")

resolvers += Resolver.bintrayRepo("hseeberger", "maven")

libraryDependencies ++= {
  val akkaV       = "2.4.11"
  Seq(
    "com.typesafe.akka" %% "akka-actor" % akkaV,
    "com.typesafe.akka" %% "akka-stream" % akkaV,
    "com.typesafe.akka" %% "akka-http-experimental" % akkaV,
    "com.typesafe.akka" %% "akka-http-spray-json-experimental" % akkaV,
    "com.typesafe.akka" %% "akka-http-testkit" % akkaV,
    "de.heikoseeberger" %% "akka-sse" % "1.11.0",
    "org.scalatest"     %% "scalatest" % "3.0.0" % "test",
    "org.scalikejdbc" %% "scalikejdbc" % "2.4.2",
    "org.scalikejdbc" %% "scalikejdbc-config" % "2.4.2",
    "mysql" % "mysql-connector-java" % "6.0.3",
    "org.slf4j" % "slf4j-simple" % "1.7.21"
  )
}

Revolver.settings
