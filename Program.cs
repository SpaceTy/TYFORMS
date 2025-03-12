using Microsoft.EntityFrameworkCore;
using TYBLOG.Data;

var builder = WebApplication.CreateBuilder(args);

// Add database context
builder.Services.AddDbContext<ApplicationDbContext>(options =>
    options.UseSqlite(builder.Configuration.GetConnectionString("DefaultConnection") ?? "Data Source=application.db"));

// Add services to the container
builder.Services.AddControllers();

// Add CORS
builder.Services.AddCors(options =>
{
    options.AddPolicy("VueCorsPolicy", builder =>
    {
        builder
            .AllowAnyHeader()
            .AllowAnyMethod()
            .AllowCredentials()
            .WithOrigins("http://localhost:5173"); // Vue dev server
    });
});

// Add OpenAPI/Swagger
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddOpenApi();

var app = builder.Build();

// Create the database if it doesn't exist
using (var scope = app.Services.CreateScope())
{
    var services = scope.ServiceProvider;
    var context = services.GetRequiredService<ApplicationDbContext>();
    context.Database.EnsureCreated();
}

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
    app.MapOpenApi();
}

app.UseHttpsRedirection();
app.UseStaticFiles();

app.UseCors("VueCorsPolicy");

app.UseRouting();
app.UseAuthorization();

app.MapControllers();

// In production, the Vue files will be served from this directory
app.UseStaticFiles();
app.MapFallbackToFile("index.html");

app.Run();
