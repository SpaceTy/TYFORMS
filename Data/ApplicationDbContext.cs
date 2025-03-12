using Microsoft.EntityFrameworkCore;
using TYBLOG.Models;

namespace TYBLOG.Data;

public class ApplicationDbContext : DbContext
{
    public ApplicationDbContext(DbContextOptions<ApplicationDbContext> options) : base(options)
    {
    }

    public DbSet<Application> Applications { get; set; } = null!;

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);
        
        // Configure entity relationships and constraints
        modelBuilder.Entity<Application>()
            .HasIndex(a => a.MinecraftUsername)
            .IsUnique();
    }
} 