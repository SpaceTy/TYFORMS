﻿// <auto-generated />
using System;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Storage.ValueConversion;
using TYBLOG.Data;

#nullable disable

namespace TYBLOG.Migrations
{
    [DbContext(typeof(ApplicationDbContext))]
    partial class ApplicationDbContextModelSnapshot : ModelSnapshot
    {
        protected override void BuildModel(ModelBuilder modelBuilder)
        {
#pragma warning disable 612, 618
            modelBuilder.HasAnnotation("ProductVersion", "9.0.2");

            modelBuilder.Entity("TYBLOG.Models.Application", b =>
                {
                    b.Property<int>("Id")
                        .ValueGeneratedOnAdd()
                        .HasColumnType("INTEGER");

                    b.Property<int>("Age")
                        .HasColumnType("INTEGER");

                    b.Property<string>("DiscordUsername")
                        .IsRequired()
                        .HasColumnType("TEXT");

                    b.Property<string>("FavoriteAboutMinecraft")
                        .IsRequired()
                        .HasColumnType("TEXT");

                    b.Property<bool>("IsReviewed")
                        .HasColumnType("INTEGER");

                    b.Property<bool>("JoinedDiscord")
                        .HasColumnType("INTEGER");

                    b.Property<string>("MinecraftUsername")
                        .IsRequired()
                        .HasColumnType("TEXT");

                    b.Property<string>("ReviewNotes")
                        .HasColumnType("TEXT");

                    b.Property<DateTime?>("ReviewedAt")
                        .HasColumnType("TEXT");

                    b.Property<DateTime>("SubmissionDate")
                        .HasColumnType("TEXT");

                    b.Property<string>("UnderstandingOfSMP")
                        .IsRequired()
                        .HasColumnType("TEXT");

                    b.HasKey("Id");

                    b.HasIndex("MinecraftUsername")
                        .IsUnique();

                    b.ToTable("Applications");
                });
#pragma warning restore 612, 618
        }
    }
}
