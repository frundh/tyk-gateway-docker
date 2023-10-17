using System.Globalization;
using System.Net;
using System.Net.Http.Json;
using System.Text.Json;
using DotNet.Testcontainers.Builders;
using Xunit;

namespace tests;

public class TykMiddlewareTests
{
  [Fact]
  public async Task Custom_headers_should_be_added_to_upstream_request()
  {
    var network = new NetworkBuilder()
      .WithName(Guid.NewGuid().ToString("D"))
      .Build();

    //var tykImage = new ImageFromDockerfileBuilder()
    //  .WithDockerfileDirectory(CommonDirectoryPath.GetGitDirectory(), string.Empty)
    //  .WithDockerfile("Dockerfile")
    //  .Build();

    //await tykImage.CreateAsync()
    //  .ConfigureAwait(false);

    var tyk = new ContainerBuilder()
      .WithName(Guid.NewGuid().ToString("D"))
      //.WithImage(tykImage.FullName)
      .WithImage("docker.tyk.io/tyk-gateway/tyk-gateway:v5.1.0")
      .WithResourceMapping(new DirectoryInfo(Path.Combine(CommonDirectoryPath.GetGitDirectory().DirectoryPath, "apps")), "/opt/tyk-gateway/apps")
      .WithResourceMapping(new DirectoryInfo(Path.Combine(CommonDirectoryPath.GetGitDirectory().DirectoryPath, "middleware")), "/opt/tyk-gateway/middleware")
      .WithResourceMapping(new FileInfo(Path.Combine(CommonDirectoryPath.GetGitDirectory().DirectoryPath, "tyk.standalone.conf")), new FileInfo("/opt/tyk-gateway/tyk.conf"))
      .WithNetwork(network)
      .WithNetworkAliases("tyk-gateway")
      .WithPortBinding(8080, true)
      .WithWaitStrategy(Wait.ForUnixContainer()
        .UntilHttpRequestIsSucceeded(request => request.ForPort(8080).ForPath("/hello").ForStatusCode(HttpStatusCode.OK))
        .UntilMessageIsLogged("Initialised API Definitions"))
      .Build();

    var redis = new ContainerBuilder()
      .WithName(Guid.NewGuid().ToString("D"))
      .WithImage("redis:6.2.7-alpine")
      .WithNetwork(network)
      .WithNetworkAliases("tyk-redis")
      .WithWaitStrategy(Wait.ForUnixContainer().UntilCommandIsCompleted("redis-cli", "ping"))
      .Build();

    var httpbin = new ContainerBuilder()
      .WithName(Guid.NewGuid().ToString("D"))
      .WithImage("kennethreitz/httpbin")
      .WithNetwork(network)
      .WithNetworkAliases("httpbin")
      .WithPortBinding(80, true)
      .WithWaitStrategy(Wait.ForUnixContainer().UntilHttpRequestIsSucceeded(request => request.ForPort(80).ForPath("/").ForStatusCode(HttpStatusCode.OK)))
      .Build();

    await network.CreateAsync()
      .ConfigureAwait(false);

    await Task.WhenAll(redis.StartAsync(), httpbin.StartAsync())
      .ConfigureAwait(false);

    await tyk.StartAsync().ConfigureAwait(false);

    var httpClient = new HttpClient();
    var requestUri = new UriBuilder(Uri.UriSchemeHttp, tyk.Hostname, tyk.GetMappedPublicPort(8080), "/keyless-test/get").Uri;

    var resp = await httpClient.GetAsync(requestUri).ConfigureAwait(false);
    var json = await resp.Content.ReadAsStringAsync();

    var data = JsonDocument.Parse(json);
    Assert.Equal("hello world", data.RootElement.GetProperty("headers").GetProperty("Custom-Header").GetString());
  }
}